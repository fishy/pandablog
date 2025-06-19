// Package blocklist provides a blocklist based on ip and user-agent regexp.
package blocklist

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/netip"
	"regexp"
	"time"

	"go.yhsif.com/ctxslog"
	"gopkg.in/yaml.v3"
)

type withRaw[T any] struct {
	raw string
	v   T
}

type matchError struct {
	ruleType string // "ip", "ua", or "uri"
	rule     string // the raw rule
	matched  string // matched ip, ua, or uri from the request
}

func (me matchError) Error() string {
	return fmt.Sprintf("%s %q matched rule %q", me.ruleType, me.matched, me.rule)
}

// A Blocklist is a IP and User-Agent based blocklist.
//
// Zero value is a valid Blocklist that matches nothing.
//
// A non-zero Blocklist must call Parse first before it can be used.
type Blocklist struct {
	Code    *int          `yaml:"code"`
	Message *string       `yaml:"message"`
	Sleep   time.Duration `yaml:"sleep"`

	IP  []string `yaml:"ip"`
	UA  []string `yaml:"ua"`
	URI []string `yaml:"uri"`

	ipPrefixes []withRaw[netip.Prefix]   `yaml:"-"`
	ua         []withRaw[*regexp.Regexp] `yaml:"-"`
	uri        []withRaw[*regexp.Regexp] `yaml:"-"`
}

// ParseYAML creates a Blocklist from yaml config.
//
// It also uses strict parsing mode.
func ParseYAML(ctx context.Context, r io.Reader) (Blocklist, error) {
	var b Blocklist
	decoder := yaml.NewDecoder(r)
	decoder.KnownFields(true)
	if err := decoder.Decode(&b); err != nil {
		return Blocklist{}, err
	}
	b.Parse(ctx)
	return b, nil
}

func parsePrefixOrSingleIP(str string) (netip.Prefix, error) {
	p, prefixErr := netip.ParsePrefix(str)
	if prefixErr == nil {
		return p, nil
	}
	ip, err := netip.ParseAddr(str)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("prefix parse error: %w, ip parse error: %w", prefixErr, err)
	}
	p = netip.PrefixFrom(ip, ip.BitLen())
	return p, nil
}

// Parse parses raw IP and UA strings.
//
// It logs all unparsable IP or UA lines as warning level logs.
func (b *Blocklist) Parse(ctx context.Context) {
	b.ipPrefixes = make([]withRaw[netip.Prefix], 0, len(b.IP))
	for _, ip := range b.IP {
		p, err := parsePrefixOrSingleIP(ip)
		if err != nil {
			slog.WarnContext(ctx, "Failed to parse ip", "err", err, "ip", ip)
			continue
		}
		b.ipPrefixes = append(b.ipPrefixes, withRaw[netip.Prefix]{
			raw: ip,
			v:   p,
		})
	}
	slog.DebugContext(ctx, "blocklist: parsed ip rules", "n", len(b.ipPrefixes))
	b.ua = make([]withRaw[*regexp.Regexp], 0, len(b.UA))
	for _, str := range b.UA {
		ua, err := regexp.Compile(str)
		if err != nil {
			slog.WarnContext(ctx, "Failed to parse user-agent regexp", "err", err, "ua", str)
			continue
		}
		b.ua = append(b.ua, withRaw[*regexp.Regexp]{
			raw: str,
			v:   ua,
		})
	}
	slog.DebugContext(ctx, "blocklist: parsed ua rules", "n", len(b.ua))
	b.uri = make([]withRaw[*regexp.Regexp], 0, len(b.URI))
	for _, str := range b.URI {
		uri, err := regexp.Compile(str)
		if err != nil {
			slog.WarnContext(ctx, "Failed to parse URI regexp", "err", err, "uri", str)
			continue
		}
		b.uri = append(b.uri, withRaw[*regexp.Regexp]{
			raw: str,
			v:   uri,
		})
	}
	slog.DebugContext(ctx, "blocklist: parsed uri rules", "n", len(b.uri))
}

// Check returns an error if the request matches the blocklist.
func (b Blocklist) Check(r *http.Request) error {
	ip := ctxslog.GCPRealIP(r)
	for _, rule := range b.ipPrefixes {
		if rule.v.Contains(ip) {
			return matchError{
				ruleType: "ip",
				rule:     rule.raw,
				matched:  ip.String(),
			}
		}
	}
	ua := r.Header.Get("user-agent")
	for _, rule := range b.ua {
		if rule.v.MatchString(ua) {
			return matchError{
				ruleType: "ua",
				rule:     rule.raw,
				matched:  ua,
			}
		}
	}
	return nil
}

// CheckAFter should be run after the normal http handler returned.
func (b Blocklist) CheckAfter(w http.ResponseWriter, r *http.Request, status int, err error) (handled bool) {
	if err != nil {
		return false
	}

	// We only want to handle 404s here
	if status != 404 {
		return false
	}

	uri := r.URL.Path
	for _, rule := range b.uri {
		if rule.v.MatchString(r.URL.Path) {
			b.writeError(r.Context(), w, matchError{
				ruleType: "uri",
				rule:     rule.raw,
				matched:  uri,
			})
			return true
		}
	}
	return false
}

// Middleware provides a HTTP middleware based on the configured blocklist.
func (b Blocklist) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/robots.txt" { // still allow them to access robots.txt
			if err := b.Check(r); err != nil {
				b.writeError(r.Context(), w, err)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (b Blocklist) writeError(ctx context.Context, w http.ResponseWriter, err error) {
	var sleep time.Duration
	if b.Sleep > 0 {
		// Sleep for b.Sleep +- 10%
		sleep = time.Duration(float64(b.Sleep) * (1 - (rand.Float64()*2-1)*0.1))
		ctx = ctxslog.Attach(ctx, "sleep", sleep)
	}
	code := http.StatusForbidden
	if b.Code != nil {
		code = *b.Code
	}
	text := http.StatusText(code)
	if b.Message != nil {
		text = *b.Message
	}
	slog.InfoContext(ctx, "blocking request", "err", err, "code", code, "text", text)
	if sleep > 0 {
		time.Sleep(sleep)
	}
	w.WriteHeader(code)
	io.WriteString(w, text)
}
