package slasher

import (
	"context"
	"testing"

	types "github.com/prysmaticlabs/eth2-types"
	dbtest "github.com/prysmaticlabs/prysm/beacon-chain/db/testing"
	slashertypes "github.com/prysmaticlabs/prysm/beacon-chain/slasher/types"
	"github.com/prysmaticlabs/prysm/shared/testutil/require"
	logTest "github.com/sirupsen/logrus/hooks/test"
)

func Test_processQueuedBlocks_DetectsDoubleProposals(t *testing.T) {
	hook := logTest.NewGlobal()
	beaconDB := dbtest.SetupDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		serviceCfg: &ServiceConfig{
			Database: beaconDB,
		},
		params:            DefaultParams(),
		beaconBlocksQueue: make([]*slashertypes.CompactBeaconBlock, 0),
	}
	currentEpochChan := make(chan types.Epoch, 0)
	exitChan := make(chan struct{})
	go func() {
		s.processQueuedBlocks(ctx, currentEpochChan)
		exitChan <- struct{}{}
	}()
	s.beaconBlocksQueue = []*slashertypes.CompactBeaconBlock{
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{1},
		},
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{1},
		},
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{1},
		},
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{2},
		},
	}
	currentEpoch := types.Epoch(4)
	currentEpochChan <- currentEpoch
	cancel()
	<-exitChan
	require.LogsContain(t, hook, "Proposer double proposal slashing")
}

func Test_processQueuedBlocks_NotSlashable(t *testing.T) {
	hook := logTest.NewGlobal()
	beaconDB := dbtest.SetupDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{
		serviceCfg: &ServiceConfig{
			Database: beaconDB,
		},
		params:            DefaultParams(),
		beaconBlocksQueue: make([]*slashertypes.CompactBeaconBlock, 0),
	}
	currentEpochChan := make(chan types.Epoch, 0)
	exitChan := make(chan struct{})
	go func() {
		s.processQueuedBlocks(ctx, currentEpochChan)
		exitChan <- struct{}{}
	}()
	s.beaconBlocksQueue = []*slashertypes.CompactBeaconBlock{
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{1},
		},
		{
			Slot:          4,
			ProposerIndex: 1,
			SigningRoot:   [32]byte{1},
		},
	}
	currentEpoch := types.Epoch(4)
	currentEpochChan <- currentEpoch
	cancel()
	<-exitChan
	require.LogsDoNotContain(t, hook, "Proposer double proposal slashing")
}
