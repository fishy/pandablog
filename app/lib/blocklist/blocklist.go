// Package blocklist provides a blocklist based on ip and user-agent regexp.
package blocklist

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"regexp"

	"go.yhsif.com/ctxslog"
	"gopkg.in/yaml.v3"
)

type withRaw[T any] struct {
	raw string
	v   T
}

type matchError struct {
	ruleType string // "ua" or "ip"
	rule     string // the raw rule
	matched  string // matched ip or ua from the request
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
	Code    *int    `yaml:"code"`
	Message *string `yaml:"message"`

	IP []string `yaml:"ip"`
	UA []string `yaml:"ua"`

	ipPrefixes []withRaw[netip.Prefix]
	ua         []withRaw[*regexp.Regexp]
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
		}
		b.ua = append(b.ua, withRaw[*regexp.Regexp]{
			raw: str,
			v:   ua,
		})
	}
	slog.DebugContext(ctx, "blocklist: parsed ua rules", "n", len(b.ua))
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

// Middleware provides a HTTP middleware based on the configured blocklist.
func (b Blocklist) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/robots.txt" { // still allow them to access robots.txt
			if err := b.Check(r); err != nil {
				slog.InfoContext(r.Context(), "blocking request", "err", err)
				code := http.StatusForbidden
				if b.Code != nil {
					code = *b.Code
				}
				w.WriteHeader(code)
				text := http.StatusText(code)
				if b.Message != nil {
					text = *b.Message
				}
				io.WriteString(w, text)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
