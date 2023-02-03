package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/hashicorp/go-sockaddr"
)

var sockaddrGetAllInterfacesGen = cel.Function("sockaddr.GetAllInterfaces",
	cel.Overload("sockaddr.GetAllInterfaces_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetAllInterfaces()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetDefaultInterfacesGen = cel.Function("sockaddr.GetDefaultInterfaces",
	cel.Overload("sockaddr.GetDefaultInterfaces_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetDefaultInterfaces()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPrivateInterfacesGen = cel.Function("sockaddr.GetPrivateInterfaces",
	cel.Overload("sockaddr.GetPrivateInterfaces_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPrivateInterfaces()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPublicInterfacesGen = cel.Function("sockaddr.GetPublicInterfaces",
	cel.Overload("sockaddr.GetPublicInterfaces_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPublicInterfaces()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrSortGen = cel.Function("sockaddr.Sort",
	cel.Overload("sockaddr.Sort_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Sort(args[0].Value().(string), args[1].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrExcludeGen = cel.Function("sockaddr.Exclude",
	cel.Overload("sockaddr.Exclude_string_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Exclude(args[0].Value().(string), args[1].Value().(string), args[2].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrIncludeGen = cel.Function("sockaddr.Include",
	cel.Overload("sockaddr.Include_string_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Include(args[0].Value().(string), args[1].Value().(string), args[2].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrAttrGen = cel.Function("sockaddr.Attr",
	cel.Overload("sockaddr.Attr_string_interface{}",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Attr(args[0].Value().(string), args[1])
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrJoinGen = cel.Function("sockaddr.Join",
	cel.Overload("sockaddr.Join_string_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Join(args[0].Value().(string), args[1].Value().(string), args[2].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrLimitGen = cel.Function("sockaddr.Limit",
	cel.Overload("sockaddr.Limit_uint_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.UintType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Limit(args[0].Value().(uint), args[1].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrOffsetGen = cel.Function("sockaddr.Offset",
	cel.Overload("sockaddr.Offset_int_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.IntType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Offset(args[0].Value().(int), args[1].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrUniqueGen = cel.Function("sockaddr.Unique",
	cel.Overload("sockaddr.Unique_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Unique(args[0].Value().(string), args[1].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrMathGen = cel.Function("sockaddr.Math",
	cel.Overload("sockaddr.Math_string_string_sockaddr.IfAddrs",

		[]*cel.Type{
			cel.StringType, cel.StringType, cel.DynType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.Math(args[0].Value().(string), args[1].Value().(string), args[2].Value().(sockaddr.IfAddrs))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPrivateIPGen = cel.Function("sockaddr.GetPrivateIP",
	cel.Overload("sockaddr.GetPrivateIP_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPrivateIP()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPrivateIPsGen = cel.Function("sockaddr.GetPrivateIPs",
	cel.Overload("sockaddr.GetPrivateIPs_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPrivateIPs()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPublicIPGen = cel.Function("sockaddr.GetPublicIP",
	cel.Overload("sockaddr.GetPublicIP_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPublicIP()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetPublicIPsGen = cel.Function("sockaddr.GetPublicIPs",
	cel.Overload("sockaddr.GetPublicIPs_",
		nil,
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetPublicIPs()
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetInterfaceIPGen = cel.Function("sockaddr.GetInterfaceIP",
	cel.Overload("sockaddr.GetInterfaceIP_string",

		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetInterfaceIP(args[0].Value().(string))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)

var sockaddrGetInterfaceIPsGen = cel.Function("sockaddr.GetInterfaceIPs",
	cel.Overload("sockaddr.GetInterfaceIPs_string",

		[]*cel.Type{
			cel.StringType,
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {

			var x SockaddrFuncs

			a0, a1 := x.GetInterfaceIPs(args[0].Value().(string))
			return types.DefaultTypeAdapter.NativeToValue([]any{
				a0, a1,
			})

		}),
	),
)
