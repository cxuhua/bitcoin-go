package script

import "errors"

var (
	SCRIPT_ERR_BAD_OPCODE                 = errors.New("SCRIPT_ERR_BAD_OPCODE")
	SCRIPT_ERR_OP_COUNT                   = errors.New("SCRIPT_ERR_OP_COUNT")
	SCRIPT_ERR_DISABLED_OPCODE            = errors.New("SCRIPT_ERR_DISABLED_OPCODE")
	SCRIPT_ERR_MINIMALDATA                = errors.New("SCRIPT_ERR_MINIMALDATA")
	SCRIPT_ERR_STACK_SIZE                 = errors.New("SCRIPT_ERR_STACK_SIZE")
	SCRIPT_ERR_UNBALANCED_CONDITIONAL     = errors.New("SCRIPT_ERR_UNBALANCED_CONDITIONAL")
	SCRIPT_ERR_INVALID_STACK_OPERATION    = errors.New("SCRIPT_ERR_INVALID_STACK_OPERATION")
	SCRIPT_ERR_NEGATIVE_LOCKTIME          = errors.New("SCRIPT_ERR_NEGATIVE_LOCKTIME")
	SCRIPT_ERR_UNSATISFIED_LOCKTIME       = errors.New("SCRIPT_ERR_UNSATISFIED_LOCKTIME")
	SCRIPT_ERR_DISCOURAGE_UPGRADABLE_NOPS = errors.New("SCRIPT_ERR_DISCOURAGE_UPGRADABLE_NOPS")
	SCRIPT_ERR_MINIMALIF                  = errors.New("SCRIPT_ERR_MINIMALIF")
	SCRIPT_ERR_VERIFY                     = errors.New("SCRIPT_ERR_VERIFY")
	SCRIPT_ERR_OP_RETURN                  = errors.New("SCRIPT_ERR_OP_RETURN")
	SCRIPT_ERR_EQUALVERIFY                = errors.New("SCRIPT_ERR_EQUALVERIFY")
	SCRIPT_ERR_NUMEQUALVERIFY             = errors.New("SCRIPT_ERR_NUMEQUALVERIFY")
	SCRIPT_ERR_SIG_FINDANDDELETE          = errors.New("SCRIPT_ERR_SIG_FINDANDDELETE")
	SCRIPT_ERR_SIG_DER                    = errors.New("SCRIPT_ERR_SIG_DER")
	SCRIPT_ERR_SIG_HIGH_S                 = errors.New("SCRIPT_ERR_SIG_HIGH_S")
	SCRIPT_ERR_SIG_HASHTYPE               = errors.New("SCRIPT_ERR_SIG_HASHTYPE")
	SCRIPT_ERR_PUBKEYTYPE                 = errors.New("SCRIPT_ERR_PUBKEYTYPE")
	SCRIPT_ERR_WITNESS_PUBKEYTYPE         = errors.New("SCRIPT_ERR_WITNESS_PUBKEYTYPE")
	SCRIPT_ERR_SIG_NULLFAIL               = errors.New("SCRIPT_ERR_SIG_NULLFAIL")
	SCRIPT_ERR_CHECKSIGVERIFY             = errors.New("SCRIPT_ERR_CHECKSIGVERIFY")
	SCRIPT_ERR_PUBKEY_COUNT               = errors.New("SCRIPT_ERR_PUBKEY_COUNT")
	SCRIPT_ERR_SIG_COUNT                  = errors.New("SCRIPT_ERR_SIG_COUNT")
	SCRIPT_ERR_SIG_NULLDUMMY              = errors.New("SCRIPT_ERR_SIG_NULLDUMMY")
	SCRIPT_ERR_CHECKMULTISIGVERIFY        = errors.New("SCRIPT_ERR_CHECKMULTISIGVERIFY")
	SCRIPT_ERR_OP_CODESEPARATOR           = errors.New("SCRIPT_ERR_OP_CODESEPARATOR")
)
