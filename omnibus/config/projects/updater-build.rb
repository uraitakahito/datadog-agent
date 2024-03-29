require "./config/projects/updater-common.rb"

# creates required build directories
dependency 'preparation'

dependency 'updater'

# version manifest file
dependency 'version-manifest'

package :xz do
end