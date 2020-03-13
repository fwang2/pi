package fs

import "os"

type And struct {
	Left, Right Node
}

type Or struct {
	Left, Right Node
}

type Not struct {
	Elem Node
}

type Node interface {
	Eval(findc *FindControl, fi os.FileInfo) bool
}

func (a And) Eval(findc *FindControl, fi os.FileInfo) bool {
	return a.Left.Eval(findc, fi) && a.Right.Eval(findc, fi)
}

func (o Or) Eval(findc *FindControl, fi os.FileInfo) bool {
	return o.Left.Eval(findc, fi) || o.Right.Eval(findc, fi)
}

func (n Not) Eval(findc *FindControl, fi os.FileInfo) bool {
	return !n.Elem.Eval(findc, fi)
}
