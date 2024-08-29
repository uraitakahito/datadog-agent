import os
import pwd
import grp
import importlib.metadata
import pkg_resources
from packaging import version

def extract_version(specifier):
    """
    Extract version from the specifier string.
    """
    try:
        # Get the first version specifier from the specifier string
        return str(next(iter(pkg_resources.Requirement.parse(f'{specifier}').specifier)))
    except Exception:
        return None

def prerm_python_installed_packages_file(directory):
    """
    Create prerm installed packages file path.
    """
    return os.path.join(directory, '.prerm_python_installed_packages.txt')

def postinst_python_installed_packages_file(directory):
    """
    Create postinst installed packages file path.
    """
    return os.path.join(directory, '.postinst_python_installed_packages.txt')

def diff_python_installed_packages_file(directory):
    """
    Create diff installed packages file path.
    """
    return os.path.join(directory, '.diff_python_installed_packages.txt')

def create_python_installed_packages_file(filename):
    """
    Create a file listing the currently installed Python dependencies.
    """
    with open(filename, 'w', encoding='utf-8') as f:
        installed_packages = importlib.metadata.distributions()
        for dist in installed_packages:
            f.write(f"{dist.metadata['Name']}=={dist.version}\n")
    os.chown(filename, pwd.getpwnam('dd-agent').pw_uid, grp.getgrnam('dd-agent').gr_gid)

def create_diff_installed_packages_file(directory):
    """
    Create a file listing the new or upgraded Python dependencies.
    """
    postinst_packages = load_requirements(postinst_python_installed_packages_file(directory))
    prerm_packages = load_requirements(prerm_python_installed_packages_file(directory))
    diff_file = diff_python_installed_packages_file(directory)
    print(f"Creating file: {diff_file}")
    with open(diff_file, 'w', encoding='utf-8') as f:
        for package_name, prerm_req in prerm_packages.items():
            postinst_req = postinst_packages.get(package_name)
            if postinst_req:
                # Extract and compare versions
                postinst_version_str = extract_version(str(postinst_req.specifier))
                prerm_version_str = extract_version(str(prerm_req.specifier))
                if postinst_version_str and prerm_version_str:
                    if version.parse(prerm_version_str) > version.parse(postinst_version_str):
                        f.write(f"{prerm_req}\n")
            else:
                # Package is new in the new file; include it
                f.write(f"{prerm_req}\n")

def install_diff_packages_file(filename):
    """
    Install all Datadog integrations and python dependencies from a file
    """
    with open(filename, 'w', encoding='utf-8') as f:
        for line in f:
            print(line.strip())

def load_requirements(filename):
    """
    Load requirements from a file.
    """
    with open(filename, 'r', encoding='utf-8') as f:
        return {req.name: req for req in pkg_resources.parse_requirements(f)}
