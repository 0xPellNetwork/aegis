import unittest
from unittest.mock import patch
import subprocess
import io
import sys

# Adjust the import path based on your project structure.
from scripts.get_previous_git_version import main, fail

class TestPreviousGitVersion(unittest.TestCase):
    """
    Unit tests for get_previous_git_version.py.
    This script accepts a version parameter and, according to the rules,
    returns an existing git tag (or a fixed value) based on the input version.
    
    Rules implemented in the script:
      1. If the input version is less than v1.1.1, output fixed "v1.0.20".
      2. If the input version is >= v1.1.1 but less than v1.1.5, output fixed "v1.1.1".
      3. If the input version is less than v1.2.0 (and ≥ v1.1.5), query for the v1.1.x series.
         E.g., input "v1.1.7" returns the highest v1.1.x tag (e.g. "v1.1.7").
      4. If the input version is between v1.2.0 (inclusive) and v2.0.0 (exclusive),
         use the series v1.(input_minor-1).x. E.g., input "v1.4" means search for v1.3.x.
      5. If the input version is v2.0.0 or higher, use the series v(passed_major-1).x.y.
    """

    def run_script_with_argument(self, argument):
        """
        Helper function that sets sys.argv with the given version argument,
        runs main(), and captures stdout, stderr, and the exit code.
        """
        original_stdout = sys.stdout
        original_stderr = sys.stderr
        original_argv = sys.argv
        try:
            fake_out = io.StringIO()
            fake_err = io.StringIO()
            sys.stdout = fake_out
            sys.stderr = fake_err
            # Simulate passing the version argument.
            sys.argv = ["get_previous_git_version.py", argument]
            exit_code = 0
            try:
                main()
            except SystemExit as e:
                exit_code = e.code if isinstance(e.code, int) else 1
            return fake_out.getvalue(), fake_err.getvalue(), exit_code
        finally:
            sys.stdout = original_stdout
            sys.stderr = original_stderr
            sys.argv = original_argv

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_no_valid_tags(self, mock_subproc):
        """
        If 'git tag --list' returns only invalid tags,
        then get_all_version_tags() returns an empty list and the script
        should fail with: "No valid version tags found (vX, vX.Y, or vX.Y.Z)."
        We trigger this by providing an input that forces the git query (e.g. "v1.1.7").
        """
        mock_subproc.return_value = "tag1\ntag2\nnot-a-version\n"
        out, err, code = self.run_script_with_argument("v1.1.7")
        print("\n--- test_no_valid_tags ---")
        print("Test Input: v1.1.7")
        print("Expected Error: No valid version tags found (vX, vX.Y, or vX.Y.Z).")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("No valid version tags found (vX, vX.Y, or vX.Y.Z).", err)
        self.assertNotEqual(code, 0)

    def test_version_less_than_v1_1_1_fixed(self):
        """
        For input versions less than v1.1.1, the script should return the fixed value "v1.0.20".
        E.g., input "v1.1" (which parses as (1,1,0)) should yield "v1.0.20".
        """
        out, err, code = self.run_script_with_argument("v1.1")
        print("\n--- test_version_less_than_v1_1_1_fixed ---")
        print("Test Input: v1.1")
        print("Expected Output: v1.0.20")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v1.0.20", out)
        self.assertEqual(code, 0)

    def test_version_less_than_v1_1_5_fixed(self):
        """
        For input versions between v1.1.1 (inclusive) and v1.1.5 (exclusive),
        the script should return the fixed value "v1.1.1".
        E.g., input "v1.1.3" yields "v1.1.1".
        """
        out, err, code = self.run_script_with_argument("v1.1.3")
        print("\n--- test_version_less_than_v1_1_5_fixed ---")
        print("Test Input: v1.1.3")
        print("Expected Output: v1.1.1")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v1.1.1", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_less_than_v1_2_with_valid_tag(self, mock_subproc):
        """
        For input versions less than v1.2.0 (and ≥ v1.1.5), the script should search for
        the v1.1.x series. For example, input "v1.1.7" should cause the script to return
        the highest tag in the v1.1.x series.
        Here, if git tags include "v1.1.7" (among others), then the output should be "v1.1.7".
        """
        mock_subproc.return_value = "v0.5\nv1.1.7\nv0.9\ninvalid\n"
        out, err, code = self.run_script_with_argument("v1.1.7")
        print("\n--- test_version_less_than_v1_2_with_valid_tag ---")
        print("Test Input: v1.1.7")
        print("Expected Output: v1.1.7")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v1.1.7", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_in_range_v1_2_to_v2_with_valid_tag(self, mock_subproc):
        """
        For input versions between v1.2.0 and v2.0.0, the script should use series
        v1.(input_minor-1).x. For example, input "v1.4" means the target series is v1.3.x.
        If the git tags include "v1.3.2" and "v1.3.4", the highest is "v1.3.4".
        """
        git_tags = "v1.0\nv1.2\nv1.3.2\nv1.3.4\nv1.5\n"
        mock_subproc.return_value = git_tags
        out, err, code = self.run_script_with_argument("v1.4")
        print("\n--- test_version_in_range_v1_2_to_v2_with_valid_tag ---")
        print("Test Input: v1.4")
        print("Expected Output: v1.3.4")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v1.3.4", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_in_range_v1_2_to_v2_missing_series(self, mock_subproc):
        """
        For input versions between v1.2.0 and v2.0.0, if no tag exists for the target series,
        the script should fail.
        For example, input "v1.4" will search for v1.3.x. If the git tags are only "v1.2" and "v1.4",
        then the script should error with "No tag found for v1.3.x".
        """
        mock_subproc.return_value = "v1.2\nv1.4\n"
        out, err, code = self.run_script_with_argument("v1.4")
        print("\n--- test_version_in_range_v1_2_to_v2_missing_series ---")
        print("Test Input: v1.4")
        print("Expected Error: No tag found for v1.3.x")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("No tag found for v1.3.x", err)
        self.assertNotEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_input_v1_2_returns_v1_1_7(self, mock_subproc):
        """
        New test case:
        If the input parameter is "v1.2" and the git tag list contains "v1.1.7",
        then the expected output is "v1.1.7".
        According to the rules, for versions ≥ v1.2.0 and < v2.0.0, the script
        uses the series v1.(input_minor-1).x. For "v1.2", this means v1.1.x.
        """
        mock_subproc.return_value = "v1.1.5\nv1.1.6\nv1.1.7\n"
        out, err, code = self.run_script_with_argument("v1.2")
        print("\n--- test_input_v1_2_returns_v1_1_7 ---")
        print("Test Input: v1.2")
        print("Expected Output: v1.1.7")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v1.1.7", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_greater_equal_v2_found_previous_major(self, mock_subproc):
        """
        For input versions greater than or equal to v2.0.0, the script should use series
        v(passed_major-1).x.y. For example, input "v3.4.2" means target major is 2.
        If git tags include "v2" (parsed as (2,0,0)) and "v2.1.3" (parsed as (2,1,3)),
        the highest is "v2.1.3".
        """
        git_tags = "v1.2\nv2\nv2.1.3\nv3\n"
        mock_subproc.return_value = git_tags
        out, err, code = self.run_script_with_argument("v3.4.2")
        print("\n--- test_version_greater_equal_v2_found_previous_major ---")
        print("Test Input: v3.4.2")
        print("Expected Output: v2.1.3")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v2.1.3", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_greater_equal_v2_not_found_previous_major(self, mock_subproc):
        """
        For input versions ≥ v2.0.0 (e.g. "v3.4.2"), if no tag exists for target series v2.x.y,
        the script should fail with "No tag found for v2.x.y".
        """
        mock_subproc.return_value = "v3\nv1.5\n"
        out, err, code = self.run_script_with_argument("v3.4.2")
        print("\n--- test_version_greater_equal_v2_not_found_previous_major ---")
        print("Test Input: v3.4.2")
        print("Expected Error: No tag found for v2.x.y")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("No tag found for v2.x.y", err)
        self.assertNotEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_greater_equal_v4_found_previous_major(self, mock_subproc):
        """
        For input "v4" (which is (4,0,0) and ≥ v2.0.0), the target series is v3.x.y.
        If git tags include "v1.5", "v2", "v3.4.1", and "v4", then the candidate is "v3.4.1".
        """
        git_tags = "v1.5\nv2\nv3.4.1\nv4\n"
        mock_subproc.return_value = git_tags
        out, err, code = self.run_script_with_argument("v4")
        print("\n--- test_version_greater_equal_v4_found_previous_major ---")
        print("Test Input: v4")
        print("Expected Output: v3.4.1")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("v3.4.1", out)
        self.assertEqual(code, 0)

    @patch("scripts.get_previous_git_version.subprocess.check_output")
    def test_version_greater_equal_v4_not_found_previous_major(self, mock_subproc):
        """
        For input "v4", the target series is v3.x.y.
        If git tags do not include any tag with major version 3,
        the script should fail with "No tag found for v3.x.y".
        """
        mock_subproc.return_value = "v4\nv2\nv1.5\n"
        out, err, code = self.run_script_with_argument("v4")
        print("\n--- test_version_greater_equal_v4_not_found_previous_major ---")
        print("Test Input: v4")
        print("Expected Error: No tag found for v3.x.y")
        print("STDOUT:", out.strip())
        print("STDERR:", err.strip())
        print("Exit Code:", code)
        self.assertIn("No tag found for v3.x.y", err)
        self.assertNotEqual(code, 0)

if __name__ == "__main__":
    unittest.main()
