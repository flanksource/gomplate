// Code generated by gencel. DO NOT EDIT.

package funcs

import "github.com/google/cel-go/cel"
import "github.com/google/cel-go/common/types"
import "github.com/google/cel-go/common/types/ref"

var netLookupIPGen = cel.Function("net.LookupIP",
	cel.Overload("net.LookupIP_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupIP(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netLookupIPsGen = cel.Function("net.LookupIPs",
	cel.Overload("net.LookupIPs_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupIPs(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netLookupCNAMEGen = cel.Function("net.LookupCNAME",
	cel.Overload("net.LookupCNAME_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupCNAME(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netLookupSRVGen = cel.Function("net.LookupSRV",
	cel.Overload("net.LookupSRV_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupSRV(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netLookupSRVsGen = cel.Function("net.LookupSRVs",
	cel.Overload("net.LookupSRVs_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupSRVs(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netLookupTXTGen = cel.Function("net.LookupTXT",
	cel.Overload("net.LookupTXT_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.LookupTXT(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netParseIPGen = cel.Function("net.ParseIP",
	cel.Overload("net.ParseIP_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.ParseIP(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netParseIPPrefixGen = cel.Function("net.ParseIPPrefix",
	cel.Overload("net.ParseIPPrefix_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.ParseIPPrefix(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netParseIPRangeGen = cel.Function("net.ParseIPRange",
	cel.Overload("net.ParseIPRange_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.ParseIPRange(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netStdParseIPGen = cel.Function("net.StdParseIP",
	cel.Overload("net.StdParseIP_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.StdParseIP(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netstdParseCIDRGen = cel.Function("net.stdParseCIDR",
	cel.Overload("net.stdParseCIDR_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.stdParseCIDR(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netStdParseCIDRGen = cel.Function("net.StdParseCIDR",
	cel.Overload("net.StdParseCIDR_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.StdParseCIDR(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCIDRHostGen = cel.Function("net.CIDRHost",
	cel.Overload("net.CIDRHost_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CIDRHost(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCidrHostGen = cel.Function("net.CidrHost",
	cel.Overload("net.CidrHost_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CidrHost(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCIDRNetmaskGen = cel.Function("net.CIDRNetmask",
	cel.Overload("net.CIDRNetmask_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CIDRNetmask(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCidrNetmaskGen = cel.Function("net.CidrNetmask",
	cel.Overload("net.CidrNetmask_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CidrNetmask(args[0])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCIDRSubnetsGen = cel.Function("net.CIDRSubnets",
	cel.Overload("net.CIDRSubnets_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CIDRSubnets(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCidrSubnetsGen = cel.Function("net.CidrSubnets",
	cel.Overload("net.CidrSubnets_interface{}_interface{}",

		[]*cel.Type{
			cel.DynType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs

			a0, a1 := x.CidrSubnets(args[0], args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCIDRSubnetSizesGen = cel.Function("net.CIDRSubnetSizes",
	cel.Overload("net.CIDRSubnetSizes_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.CIDRSubnetSizes(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var netCidrSubnetSizesGen = cel.Function("net.CidrSubnetSizes",
	cel.Overload("net.CidrSubnetSizes_interface{}",

		[]*cel.Type{
			cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x NetFuncs
			list := transferSlice[interface{}](args[0].(ref.Val))

			a0, a1 := x.CidrSubnetSizes(list...)
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
