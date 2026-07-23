#!/usr/bin/env python3
"""Validate Skill manifests against the RPC contract and example grants."""

from __future__ import annotations

import json
import re
import sys
import tomllib
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SKILLS = ROOT / "skills"
PROTOCOL = "cyberedge.v1.CyberEdge"


def fail(message: str) -> None:
    print(f"skill validation failed: {message}", file=sys.stderr)
    raise SystemExit(1)


def unique_strings(value: object, field: str, skill: str) -> list[str]:
    if not isinstance(value, list) or not value:
        fail(f"{skill}: {field} must be a non-empty array")
    if not all(isinstance(item, str) and item for item in value):
        fail(f"{skill}: {field} must contain non-empty strings")
    if len(value) != len(set(value)):
        fail(f"{skill}: {field} contains duplicates")
    return value


def main() -> None:
    proto = (ROOT / "proto/cyberedge/v1/cyberedge.proto").read_text()
    service = re.search(r"service CyberEdge \{(.*?)^}", proto, re.MULTILINE | re.DOTALL)
    if service is None:
        fail("CyberEdge service was not found in the protobuf contract")
    rpc_names = set(re.findall(r"^\s*rpc\s+(\w+)\(", service.group(1), re.MULTILINE))

    grants_document = tomllib.loads((ROOT / "config/agents.example.toml").read_text())
    grants = {grant["skill_name"]: grant for grant in grants_document.get("grants", [])}
    skill_dirs = sorted(path for path in SKILLS.iterdir() if path.is_dir())
    if not skill_dirs:
        fail("no Skills found")

    for skill_dir in skill_dirs:
        skill = skill_dir.name
        for relative in ("SKILL.md", "manifest.json", "agents/openai.yaml"):
            if not (skill_dir / relative).is_file():
                fail(f"{skill}: missing {relative}")

        manifest = json.loads((skill_dir / "manifest.json").read_text())
        if manifest.get("schema_version") != 1:
            fail(f"{skill}: schema_version must be 1")
        if manifest.get("name") != skill:
            fail(f"{skill}: manifest name must match its directory")
        if not re.fullmatch(r"\d+\.\d+\.\d+", str(manifest.get("version", ""))):
            fail(f"{skill}: version must be semantic x.y.z")
        if manifest.get("protocol") != PROTOCOL:
            fail(f"{skill}: protocol must be {PROTOCOL}")

        capabilities = unique_strings(manifest.get("capabilities"), "capabilities", skill)
        allowlist = unique_strings(manifest.get("rpc_allowlist"), "rpc_allowlist", skill)
        unknown = sorted(set(allowlist) - rpc_names)
        if unknown:
            fail(f"{skill}: unknown RPCs: {', '.join(unknown)}")

        grant = grants.get(skill)
        if grant is None:
            fail(f"{skill}: missing example grant")
        if grant.get("skill_version") != manifest["version"]:
            fail(f"{skill}: example grant version differs from manifest")
        if grant.get("capabilities") != capabilities:
            fail(f"{skill}: example grant capabilities differ from manifest")

    extra_grants = sorted(set(grants) - {path.name for path in skill_dirs})
    if extra_grants:
        fail(f"example grants reference missing Skills: {', '.join(extra_grants)}")
    print(f"validated {len(skill_dirs)} Skills against {len(rpc_names)} RPCs")


if __name__ == "__main__":
    main()
