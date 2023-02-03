// Code generated by "gencel";
// DO NOT EDIT.

package funcs

import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types/ref"

var awsEC2TagsGen = cel.Function("aws.EC2Tags",
	cel.Overload("aws.EC2Tags_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.EC2Tags()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var awsKMSEncryptGen = cel.Function("aws.KMSEncrypt",
	cel.Overload("aws.KMSEncrypt_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.KMSEncrypt(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var awsKMSDecryptGen = cel.Function("aws.KMSDecrypt",
	cel.Overload("aws.KMSDecrypt_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.KMSDecrypt(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var awsUserIDGen = cel.Function("aws.UserID",
	cel.Overload("aws.UserID_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.UserID()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var awsAccountGen = cel.Function("aws.Account",
	cel.Overload("aws.Account_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.Account()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var awsARNGen = cel.Function("aws.ARN",
	cel.Overload("aws.ARN_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x Funcs
			a0, a1 := x.ARN()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
