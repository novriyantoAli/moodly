package response

func Success(data any) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

func SuccessWithMeta(data any, meta any) Response {
	return Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
}
