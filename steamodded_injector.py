import logging
import os
import platform
import subprocess
import sys
import tempfile
from pathlib import Path

try:
    from typing import Literal
except ImportError:
    from typing_extensions import Literal

import zipfile

import requests


class SteamoddedException(Exception):
    pass


logger = logging.getLogger(__name__)
logging.basicConfig(
    level=os.environ.get("LOGLEVEL", "INFO"), format="%(levelname)s: %(message)s"
)


def download_file(url: str, output_path: Path):
    response = requests.get(url, stream=True, timeout=10)  # Add timeout argument
    response.raise_for_status()
    with output_path.open("wb") as file:
        for chunk in response.iter_content(chunk_size=8192):
            file.write(chunk)


# def decompile_lua(decompiler_path, lua_path, output_dir):
#     subprocess.run([decompiler_path, lua_path, '--output', output_dir])


def merge_directory_contents(directory_path):
    directory_content = ""
    core_file_name = "core.lua"

    if os.path.exists(directory_path):
        print(f"Processing directory: {directory_path}")

        # Process core.lua first if it exists
        core_file_path = os.path.join(directory_path, core_file_name)
        if os.path.isfile(core_file_path):
            try:
                with open(core_file_path, "r", encoding="utf-8") as file:
                    directory_content += (
                        file.read() + "\n"
                    )  # Append the core file content first
                    print(f"Appended {core_file_name} to the directory content")
            except IOError as e:
                print(f"Error reading {core_file_path}: {e}")

        # Process the rest of the .lua files
        for file_name in os.listdir(directory_path):
            if (
                file_name.endswith(".lua") and file_name != core_file_name
            ):  # Skip the core.lua file
                file_path = os.path.join(directory_path, file_name)
                try:
                    with open(file_path, "r", encoding="utf-8") as file:
                        file_content = file.read()
                        directory_content += "\n" + file_content
                        print(f"Appended {file_name} to the directory content")
                except IOError as e:
                    print(f"Error reading {file_path}: {e}")
    else:
        print(f"Directory not found: {directory_path}")
    return directory_content


def modify_main_lua(main_lua_path, base_dir, directories):
    print(f"Modifying {main_lua_path} with files from {directories} in {base_dir}")

    try:
        with open(main_lua_path, "r", encoding="utf-8") as file:
            main_lua_content = file.read()
    except IOError as e:
        print(f"Error reading {main_lua_path}: {e}")
        return

    for directory in directories:
        directory_path = os.path.join(base_dir, directory)
        print(f"Looking for directory: {directory_path}")  # Debug print
        directory_content = merge_directory_contents(directory_path)
        main_lua_content += "\n" + directory_content

    try:
        with open(main_lua_path, "w", encoding="utf-8") as file:
            file.write(main_lua_content)
    except IOError as e:
        print(f"Error writing to {main_lua_path}: {e}")


def modify_game_lua(game_lua_path):
    try:
        with open(game_lua_path, "r", encoding="utf-8") as file:
            lines = file.readlines()

        target_line = "    self.SPEEDFACTOR = 1\n"
        insert_line = "    initSteamodded()\n"  # Ensure proper indentation
        target_index = None

        for i, line in enumerate(lines):
            if target_line in line:
                target_index = i
                break  # Find the first occurrence and stop

        if target_index is not None:
            print("Target line found. Inserting new line.")
            lines.insert(target_index + 1, insert_line)
            with open(game_lua_path, "w", encoding="utf-8") as file:
                file.writelines(lines)
            print("Successfully modified game.lua.")
        else:
            print("Target line not found in game.lua.")

    except IOError as e:
        print(f"Error modifying game.lua: {e}")


def main(sfx_archive_path: Path, tmpdir: Path):
    logger.info("Starting the process...")
    logger.debug("Root working directory: %s", tmpdir)

    # Check if the SFX archive path is provided
    try:
        sfx_archive_path = Path(sys.argv[1])
        logging.info("SFX Archive received: %s", sfx_archive_path)
    except IndexError as e:
        logging.error(
            "SFX not provided. Please drag and drop the SFX archive onto this"
            " executable."
        )
        raise SteamoddedException("SFX not provided") from e

    # URL to download the LuaJIT decompiler
    # luajit_decompiler_url = ""

    # Temporary directory for operations
    # with tempfile.TemporaryDirectory() as decompiler_dir:
    # This part was used to download the LuaJit decompiler
    # luajit_decompiler_path = os.path.join(decompiler_dir, 'luajit-decompiler-v2.exe')

    # # Download LuaJIT decompiler
    # if not download_file(luajit_decompiler_url, luajit_decompiler_path):
    #     print("Failed to download LuaJIT decompiler.")
    #     sys.exit(1)

    # print("LuaJIT Decompiler downloaded.")

    # URL to download the 7-Zip suite
    seven_zip_url = "https://github.com/Steamopollys/Steamodded/raw/main/7-zip/7z.zip"

    # Temporary directory for 7-Zip suite

    # Check the operating system
    # if os.name() == 'Linux':
    #    seven_zip_path = ['wine', os.path.join(seven_zip_dir.name, "7z.exe")]
    # elif os.name == 'nt':
    #    seven_zip_path = os.path.join(seven_zip_dir.name, "7z.exe")
    # else:
    #    # Handle other operating systems or raise an error
    #    raise NotImplementedError("This script only supports Windows and Linux.")

    # Determine the operating system and prepare the 7-Zip command accordingly
    # typed 'os_name' is useful for type checking (PEP 484 explains this in detail)
    os_name: Literal["posix", "nt"] = os.name
    if os_name == "posix":
        if platform.system() == "Darwin":
            # This is macOS
            seven_zip_command = "7zz"  # Update this path as necessary for macOS
        else:
            # This is Linux or another POSIX-compliant OS
            seven_zip_command = "7zz"
    else:
        # This is for Windows
        sz_path = tmpdir / "7-Zip"
        sz_path.mkdir(exist_ok=True)
        logging.info("Downloading and extracting 7-Zip suite to %s", sz_path)

        sz_zip_path = sz_path / "7z.zip"
        try:
            download_file(seven_zip_url, sz_zip_path)
        except requests.RequestException as e:
            raise SteamoddedException(
                f"Failed to download 7-Zip suite from {seven_zip_url}"
            ) from e
        with zipfile.ZipFile(sz_zip_path, "r") as zip_ref:
            zip_ref.extractall(sz_path)
        seven_zip_command = os.path.join(sz_path, "7z.exe")

    # Check if seven_zip_command is set, and is an executable file
    try:
        subprocess.run([seven_zip_command], check=True)
    except subprocess.CalledProcessError as e:
        raise SteamoddedException(
            f"Could not locate 7-Zip executable at {seven_zip_command}"
        ) from e

    # command = seven_zip_dir + ["x", "-o" + temp_dir.name, sfx_archive_path]

    # seven_zip_path = os.path.join(seven_zip_dir.name, "7z.exe")

    # Temporary directory for extraction and modification
    workdir = tmpdir / "work"
    workdir.mkdir(exist_ok=True)

    logging.debug("Working directory: %s", workdir)
    # Extract the SFX archive
    # subprocess.run([command, "x", "-o" + temp_dir.name, sfx_archive_path])
    try:
        subprocess.run(
            [seven_zip_command, "x", f"-o{workdir.name}", sfx_archive_path], check=True
        )
        logging.info("Extraction complete.")
    except subprocess.CalledProcessError as e:
        raise SteamoddedException(
            f"Failed to extract SFX archive {sfx_archive_path} to {workdir}"
        ) from e

    # Path to main.lua and game.lua within the extracted files
    try:
        main_lua_path = workdir / "main.lua"
        assert main_lua_path.is_file(), f"main.lua not found at {main_lua_path}"
        game_lua_path = workdir / "game.lua"
        assert game_lua_path.is_file(), f"game.lua not found at {game_lua_path}"
    except AssertionError as e:
        raise SteamoddedException(e) from e

    # decompile_output_path = workdir / "output"
    # decompile_output_path.mkdir(exist_ok=True)  # Create the output directory

    # This part was used to decompile to game data
    # No longer needed
    # decompile_lua(luajit_decompiler_path, main_lua_path, decompile_output_path)
    # print("Decompilation of main.lua complete.")

    # Determine the base directory (where the .exe is located)
    if getattr(sys, "frozen", False):
        # Running in a PyInstaller or Nuitka bundle
        base_dir = os.path.dirname(sys.executable)
    else:
        # Running in a normal Python environment
        base_dir = os.path.dirname(os.path.abspath(__file__))

    # Modify main.lua
    directories = ["core", "debug", "loader"]
    modify_main_lua(main_lua_path, base_dir, directories)
    print("Modification of main.lua complete.")

    # Modify main.lua
    modify_game_lua(game_lua_path)
    print("Modification of game.lua complete.")

    # Update the SFX archive with the modified main.lua
    # subprocess.run([command, "a", sfx_archive_path, main_lua_output_path])
    subprocess.run(
        [seven_zip_command, "a", sfx_archive_path, main_lua_path], check=True
    )
    # Update the SFX archive with the modified game.lua
    # subprocess.run([command, "a", sfx_archive_path, game_lua_path])
    subprocess.run(
        [seven_zip_command, "a", sfx_archive_path, game_lua_path], check=True
    )
    print("SFX Archive updated.")

    tmpdir.cleanup()
    workdir.cleanup()

    print("Process completed successfully.")
    print("Press any key to exit...")
    input()


if __name__ == "__main__":
    try:
        with tempfile.TemporaryDirectory() as temporary_dir:
            main(temporary_dir)
    except SteamoddedException as err:
        logger.error(err)
        sys.exit(1)
