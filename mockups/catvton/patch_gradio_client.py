#!/usr/bin/env python3
"""
Patch gradio_client/utils.py to handle bool schemas.

gradio_client 1.3.0 crashes when `_json_schema_to_python_type` receives
a bool schema (e.g. `additionalProperties: true`).  The `get_type` function
does `"const" in schema` which raises TypeError for bools.
"""

import gradio_client.utils

path = gradio_client.utils.__file__
with open(path, "r") as f:
    content = f.read()

patched = False

# Patch 1: In get_type, add bool guard after function def
old1 = """def get_type(schema: dict):
    if "const" in schema:"""
new1 = """def get_type(schema: dict):
    if not isinstance(schema, dict):
        return "bool" if isinstance(schema, bool) else str(type(schema).__name__)
    if "const" in schema:"""

if old1 in content:
    content = content.replace(old1, new1, 1)
    print(f"Patched get_type in {path}")
    patched = True

# Patch 2: In _json_schema_to_python_type, add bool guard
old2 = '''def _json_schema_to_python_type(schema: Any, defs) -> str:
    """Convert the json schema into a python type hint"""
    if schema == {}:'''
new2 = '''def _json_schema_to_python_type(schema: Any, defs) -> str:
    """Convert the json schema into a python type hint"""
    if isinstance(schema, bool):
        return "Any"
    if schema == {}:'''

if old2 in content:
    content = content.replace(old2, new2, 1)
    print(f"Patched _json_schema_to_python_type in {path}")
    patched = True

if patched:
    with open(path, "w") as f:
        f.write(content)
    print("Done - patches applied")
else:
    print("WARNING: No patches applied - patterns not found")
