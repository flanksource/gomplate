package gencel

func celTypeToObj(name string) string {
	switch name {
	case "cel.StringType":
		return "String"
	case "cel.BoolType":
		return "Bool"
	case "cel.DurationType":
		return "Duration"
	case "cel.TimestampType":
		return "Timestamp"
	case "cel.IntType":
		return "Int"
	default:
		return "Unknown"
	}
}
