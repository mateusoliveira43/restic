package main

import (
	"context"

	"github.com/restic/restic/internal/backend"
	"github.com/restic/restic/internal/restic"
	"github.com/spf13/pflag"
)

type snapshotFilterOptions struct {
	Hosts []string
	Tags  restic.TagLists
	Paths []string
}

// initMultiSnapshotFilterOptions is used for commands that work on multiple snapshots
// MUST be combined with restic.FindFilteredSnapshots or FindFilteredSnapshots
func initMultiSnapshotFilterOptions(flags *pflag.FlagSet, options *snapshotFilterOptions, addHostShorthand bool) {
	hostShorthand := "H"
	if !addHostShorthand {
		hostShorthand = ""
	}
	flags.StringArrayVarP(&options.Hosts, "host", hostShorthand, nil, "only consider snapshots for this `host` (can be specified multiple times)")
	flags.Var(&options.Tags, "tag", "only consider snapshots including `tag[,tag,...]` (can be specified multiple times)")
	flags.StringArrayVar(&options.Paths, "path", nil, "only consider snapshots including this (absolute) `path` (can be specified multiple times)")
}

// initSingleSnapshotFilterOptions is used for commands that work on a single snapshot
// MUST be combined with restic.FindFilteredSnapshot
func initSingleSnapshotFilterOptions(flags *pflag.FlagSet, options *snapshotFilterOptions) {
	flags.StringArrayVarP(&options.Hosts, "host", "H", nil, "only consider snapshots for this `host`, when snapshot ID \"latest\" is given (can be specified multiple times)")
	flags.Var(&options.Tags, "tag", "only consider snapshots including `tag[,tag,...]`, when snapshot ID \"latest\" is given (can be specified multiple times)")
	flags.StringArrayVar(&options.Paths, "path", nil, "only consider snapshots including this (absolute) `path`, when snapshot ID \"latest\" is given (can be specified multiple times)")
}

// FindFilteredSnapshots yields Snapshots, either given explicitly by `snapshotIDs` or filtered from the list of all snapshots.
func FindFilteredSnapshots(ctx context.Context, be restic.Lister, loader restic.LoaderUnpacked, hosts []string, tags []restic.TagList, paths []string, snapshotIDs []string) <-chan *restic.Snapshot {
	out := make(chan *restic.Snapshot)
	go func() {
		defer close(out)
		be, err := backend.MemorizeList(ctx, be, restic.SnapshotFile)
		if err != nil {
			Warnf("could not load snapshots: %v\n", err)
			return
		}

		err = restic.FindFilteredSnapshots(ctx, be, loader, hosts, tags, paths, snapshotIDs, func(id string, sn *restic.Snapshot, err error) error {
			if err != nil {
				Warnf("Ignoring %q: %v\n", id, err)
			} else {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case out <- sn:
				}
			}
			return nil
		})
		if err != nil {
			Warnf("could not load snapshots: %v\n", err)
		}
	}()
	return out
}
