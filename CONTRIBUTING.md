# Contributing to CyberEdge

CyberEdge accepts focused changes that preserve its AI-only control plane, explicit authorization boundary, evidence chain, and strictly read-only human Web surface.

## Before opening a change

- Discuss substantial protocol, data-model, policy, or scanner changes in an Issue first.
- Never test discovery or vulnerability behavior against systems you are not explicitly authorized to assess.
- Do not add human-oriented mutation commands, mutable Web routes, arbitrary scanner flags, secret-bearing evidence collection, or multi-tenant abstractions.
- Keep generated artifacts, credentials, customer data, and scanner output out of the repository.

## Development checks

```bash
cargo fmt --check
cargo clippy --all-targets --all-features -- -D warnings
cargo test --all-targets --all-features
npm --prefix web ci
npm --prefix web run build
```

Update the Protobuf contract, machine bridge, Skill manifest, policy example, tests, and operations documentation together when a capability crosses those boundaries. Pull requests must explain authorization and evidence implications, include verification, and remain narrowly scoped.

By submitting a contribution, you agree that it is licensed under Apache License 2.0.
