from gitlab import Gitlab
from gitlab.v4.objects import Project
import os
import argparse

CI_JOB_TOKEN = os.environ["CI_JOB_TOKEN"]
GITLAB_URL = os.environ["GITLAB_URL"]

def upload(project : Project, args):
    file_path : str = args.file
    file_name = file_path.split("/")[-1]
    project.generic_packages.upload(
            package_name=args.package_name,
            package_version=args.version,
            file_name=file_name,
            path=file_path
        )
    
def download(project : Project, args):
    with open (args.output, "wb") as f:
            print("downloading..")
            project.generic_packages.download(
                package_name=args.package_name,
                package_version=args.version,
                file_name=args.file,
                action=f.write,
                streamed=True
            )

def main(args) -> None:
    gl = Gitlab(GITLAB_URL, job_token=CI_JOB_TOKEN)
    project = gl.projects.get(args.project, lazy=True)
    if args.command == "upload":
        upload(project, args)
    elif args.command == "download":
        download(project, args)
    else:
        print("Unknown command")
        


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    subparser = parser.add_subparsers(dest="command")
    subparser.required = True
    upload_parser = subparser.add_parser("upload", help="upload a package")
    upload_parser.add_argument(
        "--package-name",
        type=str,
        required=True,
        help="package name to use when uploading",
    )
    upload_parser.add_argument(
        "-f",
        "--file",
        type=str,
        required=True,
        help="file path to upload as package",
    )
    upload_parser.add_argument(
        "-p",
        "--project",
        type=str,
        required=True,
        help="ID of GitLab project for uploading packages",
    )
    upload_parser.add_argument(
        "-v",
        "--version",
        type=str,
        required=True,
        help="version to create package for",
    )
    download_parser = subparser.add_parser("download", help="download a file from a package")
    download_parser.add_argument(
        "--package-name",
        type=str,
        required=True,
        help="package name to download from",
    )
    download_parser.add_argument(
        "-f",
        "--file",
        type=str,
        required=True,
        help="file to download from package",
    )
    download_parser.add_argument(
        "-p",
        "--project",
        type=str,
        required=True,
        help="ID of GitLab project for downloading packages",
    )
    download_parser.add_argument(
        "-v",
        "--version",
        type=str,
        required=True,
        help="version to download from",
    )
    download_parser.add_argument(
        "-o",
        "--output",
        type=str,
        required=True,
        help="output file path"
    )
    args = parser.parse_args()

    main(args)
