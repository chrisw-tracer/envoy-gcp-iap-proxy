# envoy-gcp-iap-proxy
A simple Go service for Envoy external auth and GCP Identity Aware Proxy

I have been working with Google Cloud's identity aware proxy for a number of years, one of the ongoing challenges has been how 
to deal with third party applications that don't integrate with it natively, especially when programmatic access is required. 

On a recent project this created an issue specically when using a self-hosted deployment of Hashicorp Vault, a number of issues had been 
raised over the years for Vault to support GCP IAP but to no avail.

## How this solves the Vault <> GCP IAP Issue
This service works in tandem with Envoy proxy's external authz server, using Envoy as an in-cluster forward proxy for Vault.

For services like the Vault Operator, rather than be pointed directly at your Vault server's IAP address, you would deploy Envoy configured
as per the example either as a sidecar, or as a deployment of it's own. In either scenario it will need a GCP Workload Identity or Service Account JSON
with permissions to your IAP secured service.

As request come in to Envoy, the authz server contacts our simple Golang service which performs a simple `GET` request to our downstream IAP service, extracts the
`Authorization` header and passes this back to Envoy as a `Proxy-Authorization` header. The `GET` is performed when the existing token is >= 55 minutes old, or on first
request.

I'll add a diagram when I have some more time to better explain this

## Security Concerns
It should be noted, that caution needs to be excersised if using this service, you are effectively creating a mechanism to bypass Google Cloud Identity Aware Proxy, should this service ever be exposed publicly you may as well disable IAP. This should only be exposed on private networks, and where possible modified to use additional forms of authentication - mTLS or similar.
