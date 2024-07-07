# This Makefile is an easy way to run common operations.
# Execute commands like this:
# * make
# * make gcp-init
# * make gcp-push
# * make privatekey
# * make mfa
# * make passhash
# * make local-init
# * make local-run

config?=

# Load the environment variables.
include $(config).env

go=go
gcloud=gcloud --project=$(PBB_GCP_PROJECT_ID)
docker_image=$(PBB_GCP_REGION)-docker.pkg.dev/$(PBB_GCP_PROJECT_ID)/${PBB_GCP_IMAGE_NAME}/${PBB_GCP_IMAGE_NAME}
full_git_version=$(shell git rev-parse HEAD)
version_tag=$(shell echo $(full_git_version) | cut -c1-12)
build_timestamp=$(shell date +%s)

.PHONY: deploy
deploy: gcp-push

################################################################################
# Deploy application
################################################################################

.PHONY: gcp-init
gcp-init:
	@echo Pushing the initial files to Google Cloud Storage.
	gsutil mb -p $(PBB_GCP_PROJECT_ID) -l ${PBB_GCP_REGION} -c Standard gs://${PBB_GCP_BUCKET_NAME}
	gsutil versioning set on gs://${PBB_GCP_BUCKET_NAME}
	gsutil cp testdata/empty.json gs://${PBB_GCP_BUCKET_NAME}/storage/site.json
	gsutil cp testdata/empty.bin gs://${PBB_GCP_BUCKET_NAME}/storage/session.bin
	@echo Creating Artifact Registry repository on GCP
	$(gcloud) services enable artifactregistry.googleapis.com cloudbuild.googleapis.com
	$(gcloud) artifacts repositories create $(PBB_GCP_IMAGE_NAME) \
		--repository-format=docker \
		--location=$(PBB_GCP_REGION)

.PHONY: gcp-push
gcp-push: test
	@echo Pushing to Google Cloud Run.
	$(gcloud) builds submit --tag $(docker_image)
	$(gcloud) run deploy --image $(docker_image) \
		--platform managed \
		--allow-unauthenticated \
		--region ${PBB_GCP_REGION} ${PBB_GCP_CLOUDRUN_NAME} \
		--cpu 1 \
		--memory 128Mi \
		--update-env-vars VERSION_TAG=$(version_tag) \
		--update-env-vars BUILD_TIMESTAMP=$(build_timestamp) \
		--update-env-vars PBB_USERNAME=${PBB_USERNAME} \
		--update-env-vars PBB_SESSION_KEY=${PBB_SESSION_KEY} \
		--update-env-vars PBB_PASSWORD_HASH=${PBB_PASSWORD_HASH} \
		--update-env-vars PBB_MFA_KEY="${PBB_MFA_KEY}" \
		--update-env-vars PBB_GCP_PROJECT_ID=${PBB_GCP_PROJECT_ID} \
		--update-env-vars PBB_GCP_BUCKET_NAME=${PBB_GCP_BUCKET_NAME} \
		--update-env-vars PBB_ALLOW_HTML=${PBB_ALLOW_HTML}

.PHONY: privatekey
privatekey:
	@echo Generating private key for encrypting sessions.
	@echo You can paste private key this into your .env file:
	@$(go) run cmd/privatekey/main.go

.PHONY: mfa
mfa:
	@echo Generating MFA for user.
	@echo You can paste this into your .env file:
	@$(go) run cmd/mfa/main.go

.PHONY: passhash
passhash:
	@echo Generating password hash.
	@echo You can paste this into your .env file:
	@$(go) run cmd/passhash/main.go

.PHONY: local-init
local-init:
	@echo Creating session and site storage files locally.
	cp storage/initial/session.bin storage/session.bin
	cp storage/initial/site.json storage/site.json

.PHONY: local-run
local-run:
	@echo Starting local server.
	LOCALDEV=true $(go) run main.go

.PHONY: test
test:
	$(go) vet ./...
	$(go) test -race ./...
