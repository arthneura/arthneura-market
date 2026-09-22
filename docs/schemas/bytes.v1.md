# bytes.v1

id: bytes.v1 (exact, lowercase)

Opaque payload. Not CSV rules.

Chunks: fixed-size pieces of the raw bytes (CHUNK_MODE not rows).
Leaves are those pieces in order.
Empty payload rejected.
Max payload 1 MiB in v1 lab.

Checker does not interpret content.
Unknown schema is rejected.
