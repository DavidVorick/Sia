package delta

import (
	"fmt"

	"github.com/NebulousLabs/Sia/siacrypto"
	"github.com/NebulousLabs/Sia/state"
)

// Takes a ScriptInput and verifies that it's allowed to run, and then stores
// information that will prevent the script from ever being run again.
func (e *Engine) HandleScriptInput(si state.ScriptInput) {
	// If the deadline for the script has already passed, reject the
	// script.
	if si.Deadline < e.state.Metadata.Height {
		return
	}

	// If the script is 'known', it has been seen before and should not be
	// processed, therefore reject.
	if e.state.KnownScript(si) {
		return
	}

	e.Execute(si)
	e.state.LearnScript(si)
}

// TODO: add docstring
func (e *Engine) Compile(b Block) (err error) {
	// The first thing that happens is the entropy seed for the block is
	// determined. Though not implemented, this happens by pulling the latest
	// external entropy source from the block and hashing it against the germ
	// from the previous block. Right now, the germ is created but the portion
	// about external entropy is omitted.
	var externalEntropy state.Entropy // will be pulled from block
	e.state.MergeExternalEntropy(externalEntropy)

	// Next each heartbeat is iterated through and processed, checking that all
	// the vital information has been correctly assembled.
	var siblingEntropy []byte
	for i, heartbeat := range b.Heartbeats {
		// Ignore heartbeat if there's no sibling.
		if !e.state.Metadata.Siblings[i].Active() {
			continue
		}

		// Verify the signature on the heartbeat.
		verified, err := e.state.Metadata.Siblings[i].PublicKey.VerifyObject(b.HeartbeatSignatures[i], heartbeat)
		if err != nil {
			continue
		}
		if !verified {
			println("Tossing sibling for invalid signature")

			fmt.Println(e.siblingIndex)
			fmt.Println(i)
			fmt.Println(b.Height)
			fmt.Println(b.ParentBlock)
			fmt.Println("Heartbeats:")
			for i := 0; i < int(state.QuorumSize); i++ {
				fmt.Println(b.Heartbeats[i])
			}
			fmt.Println("Signatures:")
			for i := 0; i < int(state.QuorumSize); i++ {
				fmt.Println(b.HeartbeatSignatures[i])
			}
			fmt.Println("Everything else")
			fmt.Println(b.ScriptInputs)
			fmt.Println(b.UpdateAdvancements)
			fmt.Println(b.AdvancementSignatures)
			fmt.Println("Finished printing block.")

			e.state.TossSibling(byte(i))
			continue
		}

		// Verify the parent block of the heartbeat.
		if heartbeat.ParentBlock != e.state.Metadata.ParentBlock {
			println("Tossing sibling for invalid parent block")
			fmt.Println(e.siblingIndex)
			fmt.Println(i)
			fmt.Println(b.Height)
			fmt.Println(b.ParentBlock)
			fmt.Println("Heartbeats:")
			for i := 0; i < int(state.QuorumSize); i++ {
				fmt.Println(b.Heartbeats[i])
			}
			fmt.Println("Signatures:")
			for i := 0; i < int(state.QuorumSize); i++ {
				fmt.Println(b.HeartbeatSignatures[i])
			}
			fmt.Println("Everything else")
			fmt.Println(b.ScriptInputs)
			fmt.Println(b.UpdateAdvancements)
			fmt.Println(b.AdvancementSignatures)
			fmt.Println("Finished printing block.")

			e.state.TossSibling(byte(i))
			continue
		}

		// proof of storage verification

		// Append the entropy to siblingEntropy.
		siblingEntropy = append(siblingEntropy, heartbeat.Entropy[:]...)
	}

	// Hash the siblingEntropy to get the new Germ.
	e.state.Metadata.Germ = state.Entropy(siacrypto.HashBytes(siblingEntropy))

	// Process all of the script inputs. Right now, every script input is
	// processed every block, with only a few protections against inifinite loops
	// and scripting DOS attacks. The future will hold a probabilistic
	// distribution of resources based on price paid for tickets.
	for _, si := range b.ScriptInputs {
		e.HandleScriptInput(si)
	}

	// Process all of the UpdateAdvancements.
	for i, ua := range b.UpdateAdvancements {
		verified, err := e.state.Metadata.Siblings[ua.Index].PublicKey.VerifyObject(b.AdvancementSignatures[i], ua)
		if err != nil || !verified {
			continue
		}
		e.state.AdvanceUpdate(ua)
	}

	// Charge wallets for the storage they are consuming, and reward sibings for
	// the storage that is being consumed.
	e.state.ExecuteCompensation()

	// Update all passive siblings so that their PassiveWindow is reduced
	// by one.
	for i := range e.state.Metadata.Siblings {
		if !e.state.Metadata.Siblings[i].Active() && !e.state.Metadata.Siblings[i].Inactive() {
			e.state.Metadata.Siblings[i].Status -= 1
		}
	}

	// Update the metadata of the quorum.
	blockHash, err := siacrypto.HashObject(b)
	if err != nil {
		panic(err)
	}
	e.state.Metadata.ParentBlock = blockHash
	e.state.Metadata.Height++

	// Save the block.
	err = e.saveBlock(b)
	if err != nil {
		return
	}

	return
}
