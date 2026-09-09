DEFAULT_CORE = 'sy'
TITLE = 'Sessions'
ICON = 'md_layers'

import json
import os
import subprocess
import shutil
import sys
from pathlib import Path

CORE = sys.argv[1] if len(sys.argv) > 1 else DEFAULT_CORE

def run(*args):
    result = subprocess.run([CORE, *args], text=True, capture_output=True, timeout=4, check=True)
    return json.loads(result.stdout)

def segment(text, role):
    return {"text": str(text), "role": role}

def items():
    return [{"id": row["name"], "search": row["name"] + " " + row["path"],
             "segments": [segment(row["name"], "name"), segment(row["path"], "path")]} for row in run("list", "--json")]

def resolve(item):
    if not any(row["id"] == item for row in items()):
        raise ValueError("session no longer exists")
    plan = run("open", "--format", "json", item)
    if plan.get("version") != "seshy.open/v1" or plan.get("command"):
        raise ValueError("unsupported directory plan")
    if not Path(plan["cwd"]).is_dir():
        raise ValueError("session directory no longer exists")
    return {"kind": "spawn", "label": item, "cwd": plan["cwd"], "command": [], "environment": plan["environment"]}

def main():
    request = json.load(sys.stdin)
    frame = {"version": "provider/v1", "kind": "result", "requestId": request.get("requestId", "invalid")}
    try:
        if request.get("version") != "provider/v1" or request.get("kind") != "request":
            raise ValueError("expected a provider/v1 request")
        capability = request["capability"]
        if capability == "provider.validate":
            if not shutil.which(CORE):
                raise FileNotFoundError("core executable is unavailable: " + CORE)
            subprocess.run([CORE, "--help"], text=True, capture_output=True, timeout=4, check=True)
            output = {"ok": True}
        elif capability == "picker.describe":
            output = {"title": TITLE, "icon": ICON}
        elif capability == "picker.list":
            output = {"items": items()}
        elif capability == "picker.open":
            output = resolve(request["input"]["id"])
        else:
            raise ValueError("unsupported capability: " + capability)
        frame.update(status="ok", output=output)
    except (OSError, ValueError, KeyError, subprocess.SubprocessError) as error:
        frame.update(status="error", message=str(error))
    print(json.dumps(frame))

if __name__ == "__main__":
    main()
