# tsb-config-validator

the name of this package sounds much bigger than the content. However there are plans to grow it, have it in both _`container`_ and _`stand-alone`_ form (currently standalone only), add _`ldap`_ checker (currently only the ElasticSearch config is validated) and maybe to add this package to _`tctl`_ utility


### How it works:

- the package is using the users current Kubernetes context to query _`istio-system`_ namespace
- if successful it obtains:
  - Elastic Search credentials that are stored in _`elastic-credentials`_
  - Elastic Search CA certificate that are stored in _`elastic-credentials`_
  - TSB Tokens _`zipkin-token`_, _`oap-token`_, _`xcp-edge-central-auth-token`_ and _`otel-token`_
  - than TSB Control Plane CRD is read - _`telemetryStore`_ and _`managementPlane`_ sections
- When all info is collected the package analyzes the received data and tries to make an educated call to the ElasticSearch 
- if Data from Elastic search is returned then the correct config is displayed for the user to apply

### Additional checks that are done today:

- Encoded credentials might have a carriage return - it can cause unpredictable behavior (eg oap-deployment works but zipkin using the same credentials fails) - the package informs the user on any found carriage returns in the Kubernetes secret.
- CA certificate if presented by Elastic search gets placed in _`/tmp`_ directory - easy for the user to apply
- _`curl`_ command with the complete list of parameters is displayed to help with additional testing
- Tokens expiration date is validated and user is informed of any expired tokens
- Checks can be done when TSB ControlPlane is pointing to standalone ElasticSearch or via MP FrontEnvoy