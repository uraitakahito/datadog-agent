name "sds"

# default_version "4112a1c"
default_version "main"
source git: 'https://github.com/DataDog/dd-sensitive-data-scanner'

build do
    license "Apache-2.0"
    license_file "./LICENSE"

    if linux_target? || osx_target?
        command "cargo build --release", cwd: "#{project_dir}/sds-go/rust"

        if osx_target?
            command "cp sds-go/rust/target/release/libsds_go.dylib #{install_dir}/embedded/lib"
        end
        if linux_target?
            command "cp sds-go/rust/target/release/libsds_go.so #{install_dir}/embedded/lib"
        end
    end
end
