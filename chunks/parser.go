package chunks

import (
	"fmt"
)

// Parse creates an AST from a sequence of Chunks.
func Parse(chunks []Chunk) (ASTNode, error) {
	type frame struct {
		cd *ControlTagDefinition
		cn *ASTControlTag
		ap *[]ASTNode
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
		case ObjChunkType:
			*ap = append(*ap, &ASTObject{Chunk: c})
		case TextChunkType:
			*ap = append(*ap, &ASTText{Chunk: c})
		case TagChunkType:
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
					ccd, ccn = cd, &ASTControlTag{Chunk: c, cd: cd}
					*ap = append(*ap, ccn)
					ap = &ccn.body
				case cd.IsBranchTag:
					n := &ASTControlTag{Chunk: c, cd: cd}
					ccn.branches = append(ccn.branches, n)
					ap = &n.body
				case cd.IsEndTag:
					f := stack[len(stack)-1]
					ccd, ccn, ap, stack = f.cd, f.cn, f.ap, stack[:len(stack)-1]
				}
			} else if td, ok := FindTagDefinition(c.Tag); ok {
				f, err := td(c.Args)
				if err != nil {
					return nil, err
				}
				*ap = append(*ap, &ASTGenericTag{render: f})
			} else {
				return nil, fmt.Errorf("unknown tag: %s", c.Tag)
			}
			// } else if len(*ap) > 0 {
			// 	switch n := ((*ap)[len(*ap)-1]).(type) {
			// 	case *ASTChunks:
			// 		n.chunks = append(n.chunks, c)
			// 	default:
			// 		*ap = append(*ap, &ASTChunks{chunks: []Chunk{c}})
			// 	}
			// } else {
			// 	*ap = append(*ap, &ASTChunks{chunks: []Chunk{c}})
			// }
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
