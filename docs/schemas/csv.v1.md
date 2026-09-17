# csv.v1

id: csv.v1 (exact, lowercase)

UTF-8 CSV.
Row 0 header must be: company,domain,email
Rows 1..n are Merkle leaves in that order.
Max row 1024 bytes.

company: non-empty
domain: non-empty, no ://
email: contains @, not at start or end

Unknown schema is rejected.
