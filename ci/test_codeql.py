"""
This script runs a local CodeQL analysis for a specified language.

It performs the following steps:
1. Reads the CodeQL configuration from .github/codeql/codeql-config.yml.
2. Looks up the language-specific configuration.
3. Creates a CodeQL database for the project.
4. Runs all the query suites specified in the configuration file.
5. Prints a consolidated summary of all the findings at the end.

Example usage:
    python ci/test_codeql.py --language go
    python ci/test_codeql.py --language go --quiet
"""

import os
import argparse
import subprocess
import sys
from pathlib import Path
import json
from typing import List, Optional, Tuple
import yaml
from termcolor import colored

# Color constants
COLOR_SUCCESS = "green"
COLOR_ERROR = "red"
COLOR_INFO = "blue"
COLOR_HEADER = "magenta"
COLOR_HIGHLIGHT = "cyan"

# Style constants
STYLE_BOLD = "bold"

# Tag constants
TAG_TOTAL_ISSUES = "Total issues found"
TAG_NO_ISSUES = "No issues found"
TAG_ERROR = "Error"
TAG_FILE = "File"
TAG_DESCRIPTION = "Description"
TAG_ISSUE = "Issue"


def _get_workdir() -> Path:
    """
    Determines the repository's root directory from the 'WORKDIR' environment variable.

    Returns:
        Path: The resolved path to the repository root.

    Raises:
        EnvironmentError: If the 'WORKDIR' environment variable is not set.
        ValueError: If the 'WORKDIR' path is not a valid directory.
    """
    workdir = os.environ.get("WORKDIR")
    if not workdir:
        raise EnvironmentError(
            "The 'WORKDIR' environment variable is not set. "
            "Please set it to the root of the checked-out repository."
        )
    workdir = Path(workdir).resolve()
    if not workdir.is_dir():
        raise ValueError(f"WORKDIR '{workdir}' is not a valid directory.")
    return workdir


# Language configurations
LANGUAGE_CONFIG = {
    "go": {"build_command": f"{_get_workdir()}/ci/compile.sh"},
    # Add other languages here, for example:
    # "python": {"build_command": None},
}

# Path constants
CODEQL_CONFIG_PATH = f"{_get_workdir()}/.github/codeql/codeql-config.yml"


def print_issue_details(result: dict) -> None:
    """Prints the details of a single issue."""
    rule_id = result["ruleId"]
    message = result["message"]["text"]
    location = result["locations"][0]["physicalLocation"]
    file_path = location["artifactLocation"]["uri"]
    line = location.get("region", {}).get("startLine", "?")
    issue_tag = colored(f"{TAG_ISSUE}:", COLOR_INFO)
    rule_id_bold = colored(rule_id, COLOR_INFO, attrs=[STYLE_BOLD])
    print(f"\n  - {issue_tag} {rule_id_bold}")
    file_tag = colored(f"{TAG_FILE}:", COLOR_INFO)
    print(f"    {file_tag} {file_path}:{line}")
    description_tag = colored(f"{TAG_DESCRIPTION}:", COLOR_INFO)
    print(f"    {description_tag} {message}")


def print_detailed_summary(results_path: Path) -> int:
    """Prints a detailed summary of the issues found in a SARIF file."""
    try:
        with open(results_path, "r", encoding="utf-8") as f:
            data = json.load(f)

        if not data["runs"] or not data["runs"][0]["results"]:
            print(colored(f"  {TAG_NO_ISSUES}.", COLOR_SUCCESS))
            return 0

        results = data["runs"][0]["results"]
        total_issues = len(results)
        color = COLOR_ERROR if total_issues > 0 else COLOR_SUCCESS
        print(
            colored(
                f"  {TAG_TOTAL_ISSUES}: {total_issues}",
                color,
                attrs=[STYLE_BOLD],
            )
        )

        for result in results:
            print_issue_details(result)

        return total_issues

    except FileNotFoundError:
        msg = f"  {TAG_ERROR}: Results file not found at: {results_path}"
        print(colored(msg, COLOR_ERROR), file=sys.stderr)
        return 0
    except json.JSONDecodeError:
        msg = f"  {TAG_ERROR}: Failed to decode JSON from {results_path}."
        print(colored(msg, COLOR_ERROR), file=sys.stderr)
        return 0
    except (KeyError, IndexError) as e:
        msg = f"  {TAG_ERROR}: Unexpected SARIF structure in {results_path}: {e}"
        print(colored(msg, COLOR_ERROR), file=sys.stderr)
        return 0


def run_command(command: List[str], cwd: Optional[Path] = None) -> None:
    """Runs a command and exits on failure, streaming output in real-time."""
    try:
        subprocess.run(
            command,
            check=True,
            cwd=cwd,
        )
    except subprocess.CalledProcessError as e:
        print(f"Error: Command failed with exit code {e.returncode}", file=sys.stderr)
        sys.exit(e.returncode)
    except FileNotFoundError:
        print(f"Error: Command '{command[0]}' not found.", file=sys.stderr)
        print(
            "Please ensure CodeQL CLI is installed and in your PATH.",
            file=sys.stderr,
        )
        sys.exit(1)


def create_database(
    language: str,
    db_path: Path,
    workdir: Path,
    build_command: Optional[str],
    quiet: bool,
) -> None:
    """Creates a CodeQL database for a given language."""
    print(f"--- Creating CodeQL database for {language} ---")

    create_cmd = [
        "codeql",
        "database",
        "create",
        str(db_path),
        f"--language={language}",
        "--overwrite",
        "--threads=0",
        "-q" if quiet else "-v",
    ]

    if build_command:
        create_cmd.append(f"--command={build_command}")
    run_command(create_cmd, cwd=workdir)


def analyze_database(
    language: str, db_path: Path, query: str, output_path: Path, quiet: bool
) -> Tuple[str, Path]:
    """Analyzes a CodeQL database using a specific query suite."""
    print(f"--- Analyzing {language} with suite: {query} ---")
    results_path = output_path / f"results_{query}.sarif"
    verbosity = "errors" if quiet else "progress"

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
        f"--verbosity={verbosity}",
        "--threads=0",
        "--xterm-progress=auto",
        f"codeql/{language}-queries:codeql-suites/{language}-{query}.qls",
    ]
    run_command(analyze_cmd)
    return query, results_path


def run_analysis(
    workdir: Path,
    output_path: Path,
    language: str,
    build_command: Optional[str],
    quiet: bool,
) -> None:
    """Runs the entire analysis process for all configured queries."""
    with open(CODEQL_CONFIG_PATH, "r", encoding="utf-8") as f:
        codeql_config = yaml.safe_load(f)

    queries = [q["uses"] for q in codeql_config.get("queries", [])]
    all_results = []
    total_issues_found = 0

    db_path = output_path / "codeql_db"
    create_database(language, db_path, workdir, build_command, quiet)

    for query in queries:
        query, results_path = analyze_database(
            language, db_path, query, output_path, quiet
        )
        all_results.append((query, results_path))

    print(colored("\n" + "=" * 80, COLOR_HEADER))
    print(colored(" " * 25 + "CodeQL Analysis Final Summary", COLOR_HEADER))
    print(colored("=" * 80, COLOR_HEADER))

    for query, results_path in all_results:
        print(colored(f"\n--- Results for query suite: {query} ---", COLOR_HIGHLIGHT))
        total_issues_found += print_detailed_summary(results_path)

    color = COLOR_ERROR if total_issues_found > 0 else COLOR_SUCCESS
    print(colored("\n" + "=" * 80, COLOR_HEADER))
    print(
        colored(
            f"Total issues found across all suites: {total_issues_found}",
            color,
            attrs=[STYLE_BOLD],
        )
    )
    print(colored("=" * 80, COLOR_HEADER))


def main() -> None:
    """
    Parses command-line arguments and runs CodeQL analysis on a repository.
    """
    parser = argparse.ArgumentParser(
        description="Run CodeQL analysis for a specified language.",
        formatter_class=argparse.RawTextHelpFormatter,
    )
    parser.add_argument(
        "-l",
        "--language",
        required=True,
        choices=LANGUAGE_CONFIG.keys(),
        help="The programming language to analyze.",
    )
    parser.add_argument(
        "-q",
        "--quiet",
        action="store_true",
        help="Show only summary results.",
    )
    args = parser.parse_args()

    try:
        workdir = _get_workdir()
        output_path = workdir / "dist" / "codeql"
        if not output_path.is_dir():
            os.makedirs(output_path)

        language_config = LANGUAGE_CONFIG[args.language]
        build_command = language_config.get("build_command")

        run_analysis(workdir, output_path, args.language, build_command, args.quiet)
    except (ValueError, EnvironmentError) as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
