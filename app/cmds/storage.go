package cmds

import launchcmd "github.com/imfact-labs/mitum2/launch/cmd"

type Storage struct { //nolint:govet //...
	Import         ImportCommand                  `cmd:"" help:"import block data files"`
	Clean          launchcmd.CleanCommand         `cmd:"" help:"clean storage"`
	ValidateBlocks ValidateBlocksCommand          `cmd:"" help:"validate blocks in storage"`
	Status         launchcmd.StorageStatusCommand `cmd:"" help:"storage status"`
	RepairFrontier StorageRepairFrontierCommand   `cmd:"" name:"repair-frontier" help:"repair storage frontier"`
	Database       launchcmd.DatabaseCommand      `cmd:"" help:""`
}
