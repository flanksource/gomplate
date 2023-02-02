package funcs

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/hashicorp/go-sockaddr"
)

var GetAllInterfacessockaddrGen = cel.Function("GetAllInterfaces",
	cel.Overload("GetAllInterfaces_",
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

var GetDefaultInterfacessockaddrGen = cel.Function("GetDefaultInterfaces",
	cel.Overload("GetDefaultInterfaces_",
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

var GetPrivateInterfacessockaddrGen = cel.Function("GetPrivateInterfaces",
	cel.Overload("GetPrivateInterfaces_",
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

var GetPublicInterfacessockaddrGen = cel.Function("GetPublicInterfaces",
	cel.Overload("GetPublicInterfaces_",
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

var SortsockaddrGen = cel.Function("Sort",
	cel.Overload("Sort_string_sockaddr.IfAddrs",

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

var ExcludesockaddrGen = cel.Function("Exclude",
	cel.Overload("Exclude_string_string_sockaddr.IfAddrs",

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

var IncludesockaddrGen = cel.Function("Include",
	cel.Overload("Include_string_string_sockaddr.IfAddrs",

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

var AttrsockaddrGen = cel.Function("Attr",
	cel.Overload("Attr_string_interface{}",

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

var JoinsockaddrGen = cel.Function("Join",
	cel.Overload("Join_string_string_sockaddr.IfAddrs",

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

var LimitsockaddrGen = cel.Function("Limit",
	cel.Overload("Limit_uint_sockaddr.IfAddrs",

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

var OffsetsockaddrGen = cel.Function("Offset",
	cel.Overload("Offset_int_sockaddr.IfAddrs",

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

var UniquesockaddrGen = cel.Function("Unique",
	cel.Overload("Unique_string_sockaddr.IfAddrs",

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

var MathsockaddrGen = cel.Function("Math",
	cel.Overload("Math_string_string_sockaddr.IfAddrs",

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

var GetPrivateIPsockaddrGen = cel.Function("GetPrivateIP",
	cel.Overload("GetPrivateIP_",
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

var GetPrivateIPssockaddrGen = cel.Function("GetPrivateIPs",
	cel.Overload("GetPrivateIPs_",
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

var GetPublicIPsockaddrGen = cel.Function("GetPublicIP",
	cel.Overload("GetPublicIP_",
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

var GetPublicIPssockaddrGen = cel.Function("GetPublicIPs",
	cel.Overload("GetPublicIPs_",
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

var GetInterfaceIPsockaddrGen = cel.Function("GetInterfaceIP",
	cel.Overload("GetInterfaceIP_string",

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

var GetInterfaceIPssockaddrGen = cel.Function("GetInterfaceIPs",
	cel.Overload("GetInterfaceIPs_string",

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
