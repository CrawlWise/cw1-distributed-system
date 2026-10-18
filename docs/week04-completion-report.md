# Week 4 Completion Report — TCP and UDP IPC

## Project
CW1 Distributed System (`example.edu/6cc545/cw1`)

## Summary
Week 4 implemented and verified inter-node communication over TCP and UDP. The implementation extends the existing two-node Go application with a JSON message/acknowledgement protocol, TCP length-prefix framing, UDP datagrams, timeouts, cancellation, server-side validation, and an HTTP-accessible message inbox.

## Completed implementation

| Task | Work completed | Evidence/status |
| --- | --- | --- |
| 4.1 | Protocol documentation and TCP-versus-UDP comparison prepared | Confirm `docs/api-and-protocols.md` is saved in repository |
| 4.2 | `Message.Validate` and strict JSON `Decode` | Validation and malformed JSON tests passed |
| 4.3 | Four-byte big-endian TCP `WriteFrame` and `ReadFrame` | Fragmentation, limits, and incomplete-frame tests passed |
| 4.4 | `SendTCP` and `ServeTCP` | Round trip, cancellation, oversized frame rejection passed |
| 4.5 | `SendUDP` and `ServeUDP` | Round trip and malformed datagram drop passed |
| 4.6 | Two-node TCP/UDP integration | Both node inboxes returned HTTP 200 and received both protocols |
| 4.7 | Failure, timeout, and shutdown handling | Additional TCP timeout and recovery tests passed with race detection |
| 4.8 | TCP versus UDP comparison | Prepared in protocol documentation; confirm file is present |

## Protocol summary

- Message fields: `version`, `request_id`, `node_id`, `text`.
- Acknowledgement fields: `version`, `request_id`, `receiver`, `accepted`.
- Protocol version: `1`.
- Message text: 1–512 UTF-8 bytes.
- Maximum encoded payload: 4096 bytes.
- TCP: four-byte unsigned big-endian payload length followed by JSON.
- UDP: one JSON payload per datagram; no length prefix.
- Communication uses bounded deadlines and respects cancellation; no automatic retry.
- HTTP `GET /v1/messages` exposes the recent received-message inbox.

## Verified automated tests

### Week 4 IPC tests

Passed tests included:

- `TestWeek04_ServerRejectsOversizeTCP`
- `TestWeek04_UDPDropAndDeadline`
- `TestWeek04_RoundTrips` (TCP and UDP)
- `TestWeek04_CancelledClient`
- `TestWeek04_FrameFragmentation`
- `TestWeek04_FrameLimits`
- `TestWeek04_InvalidMessage`
- `TestWeek04_MessageValidationBoundaries`
- `TestWeek04_IncompleteFrames`
- `TestWeek04_TCPIncompleteFrameTimeout`
- `TestWeek04_TCPServerRecoversAfterMalformedFrame`

The last two tests were run separately with `-race` and passed. Expected log messages showed TCP incomplete-frame timeouts and invalid-frame rejection.

### Full-project race test

Command:

```powershell
go test -race ./... -count=1 2>&1 | Tee-Object -FilePath ..\evidence\week04\full-race-tests.txt
```

**Result:** PASS. Eight packages with tests passed; no failed packages and no reported data races. The command and example packages without tests were listed as `[no test files]`.

## Verified live two-node integration

Configured local endpoints:

| Node | HTTP | TCP | UDP |
| --- | --- | --- | --- |
| A | `127.0.0.1:8081` | `127.0.0.1:7001` | `127.0.0.1:7101` |
| B | `127.0.0.1:8082` | `127.0.0.1:7002` | `127.0.0.1:7102` |

- Node A inbox: HTTP 200, messages `tcp-2` and `udp-2` from `node-b`.
- Node B inbox: HTTP 200, messages `tcp-1` and `udp-1` from `node-a`.
- Both inbox outputs contained repeated request IDs. This is consistent with repeated test sends; inbox deduplication is not implemented or established by this evidence.

## TCP versus UDP

| Property | TCP | UDP |
| --- | --- | --- |
| Communication | Connection-oriented stream | Connectionless datagrams |
| Application framing | 4-byte length prefix | One JSON message per datagram |
| Delivery/order | Ordered reliable byte stream while connection is healthy | No delivery or ordering guarantee |
| Failure behavior | Connection/read/write errors and deadlines | Datagrams may be lost; client waits for acknowledgement until deadline |
| Implemented safeguards | Length limits, strict decoding, deadlines, cancellation | Datagram size limits, strict decoding, deadlines, cancellation |

Both transports use request IDs and acknowledgements, but acknowledgement does not make UDP delivery inherently reliable. No automatic retry is implemented.

## Limitations and notes

1. Repeated request IDs can appear in the inbox; no deduplication guarantee has been established.
2. UDP does not guarantee delivery, order, or retransmission.
3. Passing tests and race detection establish behavior for exercised cases; they do not prove absence of all defects.
4. Protocol documentation was prepared in the conversation; verify that the actual Markdown file is saved in `docs/api-and-protocols.md` and reviewed.
5. Git commit and remote push have not been verified by the supplied evidence.

## Evidence checklist

- [x] Week 4 IPC tests passed.
- [x] TCP incomplete-frame timeout and recovery tests passed with `-race`.
- [x] Full-project `go test -race ./... -count=1` passed after the additional tests.
- [x] Live inbox responses confirmed TCP and UDP receipt in both directions.
- [ ] Verify TCP/UDP client output files exist under `evidence/week04/`.
- [ ] Verify inbox response evidence files exist under `evidence/week04/`.
- [ ] Verify `docs/api-and-protocols.md` is saved and complete.
- [ ] Verify Git commit and push.

## Recommended checkpoint commands

From the repository root:

```powershell
git status
git diff --check
git add source/internal/ipc source/internal/state/inbox.go source/internal/app/stage.go source/cmd/ipcclient source/scripts docs/api-and-protocols.md evidence/week04
git diff --cached --stat
git commit -m "feat: complete week 4 TCP and UDP IPC implementation"
git push
```

Review `git status` before staging: other modified Week 4 files, if any, may need to be included. If the evidence directory is ignored, decide whether the coursework requires committing it.

## Next phase

Week 4 implementation and automated verification are complete. Close the documentation/evidence and Git checkpoint items, then review the Week 5 coursework requirements before beginning new implementation.
