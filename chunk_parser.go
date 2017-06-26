package main

import (
	"fmt"
)

func Parse(chunks []Chunk) (AST, error) {
	type frame struct {
		cd *ControlTagDefinition
		cn *ASTControlTag
		ap *[]AST
	}
	var (
		root  = &ASTSeq{}
		ap    = &root.Children // pointer to current node accumulation slice
		ccd   *ControlTagDefinition
		ccn   *ASTControlTag
		stack []frame // stack of control structures
	)
	for _, c := range chunks {
		switch c.Type {
		case ObjChunk:
			*ap = append(*ap, &ASTObject{chunk: c})
		case TextChunk:
			*ap = append(*ap, &ASTText{chunk: c})
		case TagChunk:
			if cd, ok := FindControlDefinition(c.Tag); ok {
				switch {
				case cd.RequiresParent() && !cd.CompatibleParent(ccd):
					suffix := ""
					if ccd != nil {
						suffix = "; immediate parent is " + ccd.Name
					}
					return nil, fmt.Errorf("%s not inside %s%s", cd.Name, cd.Parent.Name, suffix)
				case cd.IsStartTag():
					stack = append(stack, frame{cd: ccd, cn: ccn, ap: ap})
					ccd, ccn = cd, &ASTControlTag{chunk: c, cd: cd}
					*ap = append(*ap, ccn)
					ap = &ccn.body
				case cd.IsBranchTag:
					n := &ASTControlTag{chunk: c, cd: cd}
					ccn.branches = append(ccn.branches, n)
					ap = &n.body
				case cd.IsEndTag:
					f := stack[len(stack)-1]
					ccd, ccn, ap, stack = f.cd, f.cn, f.ap, stack[:len(stack)-1]
				}
			} else if len(*ap) > 0 {
				switch n := ((*ap)[len(*ap)-1]).(type) {
				case *ASTChunks:
					n.chunks = append(n.chunks, c)
				default:
					*ap = append(*ap, &ASTChunks{chunks: []Chunk{c}})
				}
			} else {
				*ap = append(*ap, &ASTChunks{chunks: []Chunk{c}})
			}
		}
	}
	if ccd != nil {
		return nil, fmt.Errorf("unterminated %s tag", ccd.Name)
	}
	if len(root.Children) == 1 {
		return root.Children[0], nil
	}
	return root, nil
}
