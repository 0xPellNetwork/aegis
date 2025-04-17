import unittest
import io
import sys
from scripts.get_upgrade_version import main  # Adjust this import if needed

class TestGetUpgradeVersion(unittest.TestCase):
    def run_script_with_arg(self, arg):
        """
        Helper method to run main() with the given command-line argument.
        Temporarily override sys.argv and capture stdout, stderr, and the exit code.
        """
        backup_stdout = sys.stdout
        backup_stderr = sys.stderr
        backup_argv = sys.argv
        try:
            fake_out = io.StringIO()
            fake_err = io.StringIO()
            sys.stdout = fake_out
            sys.stderr = fake_err
            # Simulate command-line arguments: script name and version argument.
            sys.argv = ["script.py", arg]
            exit_code = 0
            try:
                main()
            except SystemExit as e:
                exit_code = e.code if isinstance(e.code, int) else 1
            out = fake_out.getvalue()
            err = fake_err.getvalue()
            return out, err, exit_code
        finally:
            sys.stdout = backup_stdout
            sys.stderr = backup_stderr
            sys.argv = backup_argv

    def run_script_without_arg(self):
        """
        Helper method to run main() without any version argument.
        """
        backup_stdout = sys.stdout
        backup_stderr = sys.stderr
        backup_argv = sys.argv
        try:
            fake_out = io.StringIO()
            fake_err = io.StringIO()
            sys.stdout = fake_out
            sys.stderr = fake_err
            # Simulate running the script without any argument.
            sys.argv = ["script.py"]
            exit_code = 0
            try:
                main()
            except SystemExit as e:
                exit_code = e.code if isinstance(e.code, int) else 1
            out = fake_out.getvalue()
            err = fake_err.getvalue()
            return out, err, exit_code
        finally:
            sys.stdout = backup_stdout
            sys.stderr = backup_stderr
            sys.argv = backup_argv

    def test_missing_argument(self):
        """
        When no version argument is provided, the script should print a usage error.
        """
        out, err, code = self.run_script_without_arg()
        test_input = "No Argument"
        expected = "Usage: script.py <version>"
        print("\n--- test_missing_argument ---")
        print("Test Input:", test_input)
        print("Expected Output (error contains):", expected)
        print("Actual STDOUT:", out.strip())
        print("Actual STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn("Usage:", err)
        self.assertNotEqual(code, 0)

    def test_invalid_version_format(self):
        """
        When an invalid version string is provided, the script should print an error message.
        """
        test_input = "invalid"
        expected = "Invalid version format"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_invalid_version_format ---")
        print("Test Input:", test_input)
        print("Expected Output (error contains):", expected)
        print("Actual STDOUT:", out.strip())
        print("Actual STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn("Invalid version format", err)
        self.assertNotEqual(code, 0)

    def test_version_less_than_v1_1_1(self):
        """
        If the version is less than v1.1.1 (e.g., v1.0.50), the script should output "v1.0.20".
        """
        test_input = "v1.0.50"
        expected = "v1.0.20"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_version_less_than_v1_1_1 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertEqual(code, 0)

    def test_version_between_v1_1_1_and_v1_1_5(self):
        """
        If the version is between v1.1.1 (inclusive) and v1.1.5 (exclusive), e.g., v1.1.3,
        the script should output "v1.1.1".
        """
        test_input = "v1.1.3"
        expected = "v1.1.1"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_version_between_v1_1_1_and_v1_1_5 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertEqual(code, 0)

    def test_version_between_v1_1_5_and_v1_2(self):
        """
        If the version is between v1.1.5 (inclusive) and v1.2.0 (exclusive), e.g., v1.1.7,
        the script should output "v1.1.5".
        """
        test_input = "v1.1.7"
        expected = "v1.1.5"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_version_between_v1_1_5_and_v1_2 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertEqual(code, 0)

    def test_version_in_range_v1_2_to_v2(self):
        """
        If the version is between v1.2.0 (inclusive) and v2.0.0 (exclusive), e.g., v1.2.3,
        the script should output the major and minor version, e.g., "v1.2".
        """
        test_input = "v1.2.3"
        expected = "v1.2"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_version_in_range_v1_2_to_v2 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        # Make sure the patch version is not included in the output.
        self.assertNotIn("v1.2.3", out)
        self.assertEqual(code, 0)

    def test_exactly_v1_2(self):
        """
        If the version is exactly v1.2 (parsed as v1.2.0), the script should output "v1.2".
        """
        test_input = "v1.2"
        expected = "v1.2"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_exactly_v1_2 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertEqual(code, 0)

    def test_version_greater_equal_v2(self):
        """
        If the version is v2.4.5 (>= v2.0.0), the script should output only the major version ("v2").
        """
        test_input = "v2.4.5"
        expected = "v2"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_version_greater_equal_v2 ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertNotIn("v2.4.5", out)
        self.assertEqual(code, 0)

    def test_larger_major_version(self):
        """
        If the version is v3.4.2 (>= v2.0.0), the script should output only the major version ("v3").
        """
        test_input = "v3.4.2"
        expected = "v3"
        out, err, code = self.run_script_with_arg(test_input)
        print("\n--- test_larger_major_version ---")
        print("Test Input:", test_input)
        print("Expected Output:", expected)
        print("Actual Output:", out.strip())
        print("STDERR:", err.strip())
        print("EXIT CODE:", code)
        self.assertIn(expected, out)
        self.assertNotIn("v3.4.2", out)
        self.assertEqual(code, 0)

if __name__ == "__main__":
    unittest.main()
