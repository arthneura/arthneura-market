# Schemas

The market only accepts these ids. Exact lowercase. Empty and unknown are 400.

csv.v1 is the original lab object: UTF-8 CSV, header
`company,domain,email`, one Merkle leaf per row.

bytes.v1 is a file with no interpretation. api.v1 is a canned response
body, not a live call. job.v1 is a build artifact. meter.v1 is a usage
receipt. The last four use the same byte-chunk path.

A settled commitment means the bytes that were committed were delivered
and the consumer closed. It does not mean the CSV is true, the job ran,
the HTTP service exists, or the meter counted honestly.

Happy paths:

    scripts/market-csv-settle.sh
    scripts/market-bytes-settle.sh
    scripts/market-api-settle.sh
    scripts/market-job-settle.sh
    scripts/market-meter-settle.sh

MCP lists the bazaar. It does not sign. See ../mcp-reads.md.
