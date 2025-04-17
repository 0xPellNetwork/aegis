#!/usr/bin/env python3
import sys
import re
import subprocess

def fail(msg: str):
    """
    Print an error message to stderr and exit with code 1.
    """
    print(f"Error: {msg}", file=sys.stderr)
    sys.exit(1)

def parse_major_minor_patch(tag: str):
    """
    Parse version tags of the form vX, vX.Y, or vX.Y.Z.
    Returns a tuple (major, minor, patch) if the tag matches,
    otherwise returns None.
    """
    # Three-segment format: vX.Y.Z
    match = re.match(r"^v(\d+)\.(\d+)\.(\d+)$", tag)
    if match:
        major = int(match.group(1))
        minor = int(match.group(2))
        patch = int(match.group(3))
        return (major, minor, patch)
    
    # Two-segment format: vX.Y (assume patch = 0)
    match = re.match(r"^v(\d+)\.(\d+)$", tag)
    if match:
        major = int(match.group(1))
        minor = int(match.group(2))
        return (major, minor, 0)
    
    # One-segment format: vX (assume minor = 0, patch = 0)
    match = re.match(r"^v(\d+)$", tag)
    if match:
        major = int(match.group(1))
        return (major, 0, 0)
    
    return None

def get_all_version_tags():
    """
    Retrieve all git tags that match the format vX, vX.Y, or vX.Y.Z,
    and return a list of tuples: [(tag_name, major, minor, patch), ...].
    """
    try:
        output = subprocess.check_output(["git", "tag", "--list"], encoding="utf-8")
    except subprocess.CalledProcessError as e:
        fail(f"Failed to list git tags: {e}")
    
    lines = output.strip().split("\n")
    results = []
    for t in lines:
        version_info = parse_major_minor_patch(t)
        if version_info is not None:
            results.append((t, version_info[0], version_info[1], version_info[2]))
    return results

def find_largest_tag(tags):
    """
    Given a list of (tag_name, major, minor, patch), sort them in ascending order
    by (major, minor, patch) and return the last (largest) entry.
    Returns None if the list is empty.
    """
    if not tags:
        return None
    tags.sort(key=lambda x: (x[1], x[2], x[3]))
    return tags[-1]

def find_largest_in_major_minor(tags, major, minor):
    """
    Given a list of tags, find the one with the specified major and minor version,
    and the largest patch version.
    Returns None if no matching tag is found.
    """
    candidates = [x for x in tags if x[1] == major and x[2] == minor]
    if not candidates:
        return None
    candidates.sort(key=lambda x: x[3])  # sort by patch ascending
    return candidates[-1]

def find_largest_in_major(tags, major):
    """
    Given a list of tags, find the one with the specified major version,
    and the largest (minor, patch) combination.
    Returns None if no matching tag is found.
    """
    candidates = [x for x in tags if x[1] == major]
    if not candidates:
        return None
    candidates.sort(key=lambda x: (x[2], x[3]))  # sort by (minor, patch) ascending
    return candidates[-1]

def main():
    """
    This script accepts a version string as a command-line argument and returns
    an existing git tag according to the following rules:
    
    1. If the passed version is less than v1.1.1, output fixed value "v1.0.20".
    2. If the passed version is less than v1.1.5, output fixed value "v1.1.1".
    3. If the passed version is less than v1.2.0, query git tags for the v1.1.x series
       and return the tag with the highest patch version.
    4. If the passed version is between v1.2.0 (inclusive) and v2.0.0 (exclusive),
       subtract 1 from the minor version (i.e. use series v1.(passed_minor-1).x) and return
       the tag with the highest patch version.
    5. If the passed version is v2.0.0 or greater, subtract 1 from the major version (i.e.
       use series v(passed_major-1).x.y) and return the tag with the largest (minor, patch).
       
    If no matching git tag is found, the script exits with an error.
    """
    if len(sys.argv) != 2:
        fail("Usage: get_previous_git_version.py <version>")
    
    input_version = sys.argv[1]
    parsed = parse_major_minor_patch(input_version)
    if parsed is None:
        fail("Invalid version format. Use vX, vX.Y, or vX.Y.Z.")
    
    p_major, p_minor, p_patch = parsed

    # Rule 1: If version < v1.1.1, return fixed "v1.0.20"
    if (p_major, p_minor, p_patch) < (1, 1, 1):
        print("v1.0.20")
        return

    # Rule 2: If version < v1.1.5, return fixed "v1.1.1"
    if (p_major, p_minor, p_patch) < (1, 1, 5):
        print("v1.1.1")
        return

    # Retrieve all valid git tags.
    tags = get_all_version_tags()
    if not tags:
        fail("No valid version tags found (vX, vX.Y, or vX.Y.Z).")
    
    # Rule 3: If version < v1.2.0, use the v1.1.x series.
    if (p_major, p_minor, p_patch) < (1, 2, 0):
        target_major = 1
        target_minor = 1
        candidate = find_largest_in_major_minor(tags, target_major, target_minor)
        if candidate is None:
            fail(f"No tag found for v{target_major}.{target_minor}.x")
        print(candidate[0])
        return

    # Rule 4: If v1.2.0 <= version < v2.0.0, use series v1.(p_minor-1).x.
    if (p_major, p_minor, p_patch) < (2, 0, 0):
        target_major = 1
        target_minor = p_minor - 1
        if target_minor < 0:
            fail(f"Invalid minor version derived from parameter: {input_version}")
        candidate = find_largest_in_major_minor(tags, target_major, target_minor)
        if candidate is None:
            fail(f"No tag found for v{target_major}.{target_minor}.x")
        print(candidate[0])
        return

    # Rule 5: If version >= v2.0.0, use series v(p_major-1).x.y.
    target_major = p_major - 1
    if target_major < 1:
        fail("Logic error: target major is below 1.")
    candidate = find_largest_in_major(tags, target_major)
    if candidate is None:
        fail(f"No tag found for v{target_major}.x.y")
    print(candidate[0])
    return

if __name__ == "__main__":
    main()
