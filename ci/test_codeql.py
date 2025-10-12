"""
This script runs a local CodeQL analysis for the Go language.

It performs the following steps:
1. Reads the CodeQL configuration from .github/codeql/codeql-config.yml.
2. Creates a CodeQL database for the Go project.
3. Runs all the query suites specified in the configuration file.
4. Prints a consolidated summary of all the findings at the end.
"""

import os
import subprocess
import sys
from pathlib import Path
import json
import yaml
from termcolor import colored


class CodeQLAnalyzer:
    """Encapsulates the logic for running CodeQL analysis."""

    def __init__(self, workdir):
        """Initializes the analyzer with the working directory."""
        if not workdir:
            raise ValueError("WORKDIR environment variable is not set.")
        self.repo_path = Path(workdir).resolve()
        if not self.repo_path.is_dir():
            raise ValueError(f"WORKDIR '{self.repo_path}' is not a valid directory.")
        self.output_path = self.repo_path / "dist" / "codeql"
        if not self.output_path.is_dir():
            os.makedirs(self.output_path)

    def _print_detailed_summary(self, results_path):
        """Prints a detailed summary of the issues found in a SARIF file."""
        try:
            with open(results_path, "r", encoding="utf-8") as f:
                data = json.load(f)

            if not data["runs"] or not data["runs"][0]["results"]:
                print(colored("  No issues found.", "green"))
                return 0

            results = data["runs"][0]["results"]
            total_issues = len(results)
            color = "red" if total_issues > 0 else "green"
            print(
                colored(f"  Total issues found: {total_issues}", color, attrs=["bold"])
            )

            for result in results:
                rule_id = result["ruleId"]
                message = result["message"]["text"]
                location = result["locations"][0]["physicalLocation"]
                file_path = location["artifactLocation"]["uri"]
                line = location.get("region", {}).get("startLine", "?")
                print(
                    f"\n  - {colored('Issue:', 'blue')} "
                    f"{colored(rule_id, 'blue', attrs=['bold'])}"
                )
                print(f"    {colored('File:', 'blue')} {file_path}:{line}")
                print(f"    {colored('Description:', 'blue')} {message}")

            return total_issues

        except FileNotFoundError:
            print(
                colored(f"  Error: Results file not found at: {results_path}", "red"),
                file=sys.stderr,
            )
            return 0
        except Exception as e:
            print(
                colored(f"  An error occurred while printing the summary: {e}", "red"),
                file=sys.stderr,
            )
            return 0

    def _run_command(self, command, cwd=None):
        """Runs a command and exits on failure, streaming output in real-time."""
        try:
            subprocess.run(
                command,
                check=True,
                cwd=cwd,
            )
        except subprocess.CalledProcessError as e:
            print(f"Command failed with exit code {e.returncode}", file=sys.stderr)
            sys.exit(e.returncode)
        except FileNotFoundError:
            print(f"Error: Command '{command[0]}' not found.", file=sys.stderr)
            print(
                "Please ensure CodeQL CLI is installed and in your PATH.",
                file=sys.stderr,
            )
            sys.exit(1)

    def _create_database(self, language, db_path):
        """Creates a CodeQL database for a given language."""
        print(f"--- Creating CodeQL database for {language} ---")
        create_cmd = [
            "codeql",
            "database",
            "create",
            str(db_path),
            f"--language={language}",
            "--overwrite",
        ]
        if language == "go":
            create_cmd.append("--command=/opt/ci/compile.sh")
        self._run_command(create_cmd, cwd=self.repo_path)

    def _analyze_database(self, language, db_path, query):
        """Analyzes a CodeQL database using a specific query suite."""
        qls_path = f"codeql/{language}-queries:codeql-suites/{language}-{query}.qls"
        print(f"--- Analyzing {language} with suite: {query} ---")
        results_path = self.output_path / f"results_{query}.sarif"

        analyze_cmd = [
            "codeql",
            "database",
            "analyze",
            str(db_path),
            "--format=sarif-latest",
            f"--output={results_path}",
            f"--search-path={os.environ.get('CODEQL_SEARCH_PATH')}",
            "--rerun",
            "--sarif-add-baseline-file-info",
            "--sarif-add-snippets",
            "--verbosity=progress",
            "--threads=0",
            "--xterm-progress=auto",
            qls_path,
        ]
        self._run_command(analyze_cmd)
        return query, results_path

    def run(self):
        """Runs the entire analysis process for all configured queries."""
        config_path = self.repo_path / ".github/codeql/codeql-config.yml"
        with open(config_path, "r", encoding="utf-8") as f:
            codeql_config = yaml.safe_load(f)

        queries = [q["uses"] for q in codeql_config.get("queries", [])]
        language = "go"
        all_results = []
        total_issues_found = 0

        db_path = self.output_path / "codeql_db"
        self._create_database(language, db_path)

        for query in queries:
            query, results_path = self._analyze_database(language, db_path, query)
            all_results.append((query, results_path))

        print(colored("\n" + "=" * 80, "magenta"))
        print(colored(" " * 25 + "CodeQL Analysis Final Summary", "magenta"))
        print(colored("=" * 80, "magenta"))

        for query, results_path in all_results:
            print(colored(f"\n--- Results for query suite: {query} ---", "cyan"))
            total_issues_found += self._print_detailed_summary(results_path)

        color = "red" if total_issues_found > 0 else "green"
        print(colored("\n" + "=" * 80, "magenta"))
        print(
            colored(
                f"Total issues found across all suites: {total_issues_found}",
                color,
                attrs=["bold"],
            )
        )
        print(colored("=" * 80, "magenta"))


def main():
    """
    Runs CodeQL analysis on a repository.
    """
    try:
        workdir = os.environ.get("WORKDIR")
        analyzer = CodeQLAnalyzer(workdir)
        analyzer.run()
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
