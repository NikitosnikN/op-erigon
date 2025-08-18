package jsonrpc

import (
	"github.com/erigontech/erigon-lib/log/v3"
	libcommon "github.com/erigontech/erigon-lib/common"
	"github.com/erigontech/erigon/core/vm"
	"github.com/holiman/uint256"
)

type OverlayCreateTracer struct {
	contractAddress libcommon.Address
	isCapturing     bool
	code            []byte
	gasCap          uint64
	err             error
	resultCode      []byte
	evm             *vm.EVM
}

// Transaction level
func (ct *OverlayCreateTracer) CaptureTxStart(gasLimit uint64) {}
func (ct *OverlayCreateTracer) CaptureTxEnd(restGas uint64)    {}

// Top call frame
func (ct *OverlayCreateTracer) CaptureStart(env *vm.EVM, from libcommon.Address, to libcommon.Address, precompile bool, create bool, input []byte, gas uint64, value *uint256.Int, code []byte) {
	ct.evm = env
	log.Debug("OverlayCreateTracer CaptureStart", 
		"from", from.Hex(), 
		"to", to.Hex(), 
		"create", create, 
		"targetAddr", ct.contractAddress.Hex(),
		"matches", to == ct.contractAddress)
	
	// If this is the top-level CREATE call for our target contract
	if create && to == ct.contractAddress && !ct.isCapturing {
		log.Debug("OverlayCreateTracer: Starting capture for top-level CREATE")
		ct.isCapturing = true
		_, _, _, err := ct.evm.OverlayCreate(vm.AccountRef(from), vm.NewCodeAndHash(ct.code), ct.gasCap, value, to, vm.CREATE, true /* incrementNonce */)
		if err != nil {
			log.Debug("OverlayCreateTracer: OverlayCreate error", "err", err)
			ct.err = err
		} else {
			ct.resultCode = ct.evm.IntraBlockState().GetCode(ct.contractAddress)
			log.Debug("OverlayCreateTracer: Captured code", "codeLen", len(ct.resultCode))
		}
	}
}
func (ct *OverlayCreateTracer) CaptureEnd(output []byte, usedGas uint64, err error) {
	log.Debug("OverlayCreateTracer CaptureEnd", "outputLen", len(output), "err", err)
}

// Rest of the frames
func (ct *OverlayCreateTracer) CaptureEnter(typ vm.OpCode, from libcommon.Address, to libcommon.Address, precompile bool, create bool, input []byte, gas uint64, value *uint256.Int, code []byte) {
	log.Debug("OverlayCreateTracer CaptureEnter", 
		"opcode", typ.String(),
		"from", from.Hex(), 
		"to", to.Hex(), 
		"create", create, 
		"targetAddr", ct.contractAddress.Hex(),
		"matches", to == ct.contractAddress,
		"isCapturing", ct.isCapturing)
	
	if ct.isCapturing {
		return
	}

	if create && to == ct.contractAddress {
		log.Debug("OverlayCreateTracer: Starting capture for CREATE2/internal CREATE")
		ct.isCapturing = true
		_, _, _, err := ct.evm.OverlayCreate(vm.AccountRef(from), vm.NewCodeAndHash(ct.code), ct.gasCap, value, to, typ, true /* incrementNonce */)
		if err != nil {
			log.Debug("OverlayCreateTracer: OverlayCreate error in CaptureEnter", "err", err)
			ct.err = err
		} else {
			ct.resultCode = ct.evm.IntraBlockState().GetCode(ct.contractAddress)
			log.Debug("OverlayCreateTracer: Captured code in CaptureEnter", "codeLen", len(ct.resultCode))
		}
	}
}
func (ct *OverlayCreateTracer) CaptureExit(output []byte, usedGas uint64, err error) {
	log.Debug("OverlayCreateTracer CaptureExit", "outputLen", len(output), "err", err)
}

// Opcode level
func (ct *OverlayCreateTracer) CaptureState(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, rData []byte, depth int, err error) {
}
func (ct *OverlayCreateTracer) CaptureFault(pc uint64, op vm.OpCode, gas, cost uint64, scope *vm.ScopeContext, depth int, err error) {
}
