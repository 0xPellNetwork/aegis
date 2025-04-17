#!/usr/bin/env python3
import sys
import re

def fail(msg: str):
    """
    Print an error message to stderr and exit the program with status 1.
    """
    print(f"Error: {msg}", file=sys.stderr)
    sys.exit(1)

def parse_version(tag: str):
    """
    Parse a version string in the format 'vX', 'vX.Y', or 'vX.Y.Z'.
    Returns a tuple (major, minor, patch) if the format is valid.
    Returns None if the version string does not match any of the patterns.
    """
    # Match full version: vX.Y.Z
    match_full = re.match(r"^v(\d+)\.(\d+)\.(\d+)$", tag)
    if match_full:
        major = int(match_full.group(1))
        minor = int(match_full.group(2))
        patch = int(match_full.group(3))
        return (major, minor, patch)

    # Match two-part version: vX.Y (assume patch = 0)
    match_two = re.match(r"^v(\d+)\.(\d+)$", tag)
    if match_two:
        major = int(match_two.group(1))
        minor = int(match_two.group(2))
        return (major, minor, 0)

    # Match one-part version: vX (assume minor = 0, patch = 0)
    match_one = re.match(r"^v(\d+)$", tag)
    if match_one:
        major = int(match_one.group(1))
        return (major, 0, 0)

    return None

def main():
    # Ensure exactly one command-line argument is provided (the version string)
    if len(sys.argv) != 2:
        fail("Usage: script.py <version> (e.g., script.py v1.2.3)")
    
    version_input = sys.argv[1]
    parsed = parse_version(version_input)
    if parsed is None:
        fail("Invalid version format. Please use formats like v1, v1.2, or v1.2.3")
    
    major, minor, patch = parsed
    version_tuple = (major, minor, patch)
    
    # Determine the output based on the version thresholds:
    #
    # 1. If version < v1.1.1, print "v1.0.20"
    if version_tuple < (1, 1, 1):
        print("v1.0.20")
    # 2. If v1.1.1 <= version < v1.1.5, print "v1.1.1"
    elif version_tuple < (1, 1, 5):
        print("v1.1.1")
    # 3. If v1.1.5 <= version < v1.2.0, print "v1.1.5"
    elif version_tuple < (1, 2, 0):
        print("v1.1.5")
    # 4. If v1.2.0 <= version < v2.0.0, print "v{major}.{minor}"
    elif version_tuple < (2, 0, 0):
        print(f"v{major}.{minor}")
    # 5. If version >= v2.0.0, print "v{major}"
    else:
        print(f"v{major}")

if __name__ == "__main__":
    main()
