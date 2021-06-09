package cmap

//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMap(string,interface{},KeyHashStr)"
//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMapUint32Set(uint32,struct{},KeyHashUint32)"
//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMapUint32Uint32(uint32,uint32,KeyHashUint32)"
//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMapUint32Uint64(uint32,uint64,KeyHashUint32)"
//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMapStringUint64(string,uint64,KeyHashStr)"
//go:generate gotemplate -outfmt gen_%v "github.com/funbytes/modern-go/container/templates/cmap" "ConcurrentMapStringString(string,string,KeyHashStr)"
