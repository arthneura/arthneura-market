# api.v1

id: api.v1 (exact, lowercase)

Lab v1: the sold object is a response body, not a live third-party call.

Payload = raw bytes of that body.
Chunking = same as bytes.v1 (default CHUNK_MODE=bytes).
Empty body rejected.

v1 does not require JSON. A later checker may require a status field.
Unknown schema is rejected.
