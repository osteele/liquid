package render

// Compile parses a source template. It returns an AST root, that can be evaluated.
func (s Config) Compile(source string) (ASTNode, error) {
	root, err := s.Parse(source)
	if err != nil {
		return nil, err
	}
	return s.compileNode(root)
}

// nolint: gocyclo
func (s Config) compileNode(n ASTNode) (ASTNode, error) {
	switch n := n.(type) {
	case *ASTBlock:
		for _, child := range n.Body {
			if _, err := s.compileNode(child); err != nil {
				return nil, err
			}
		}
		for _, branch := range n.Branches {
			if _, err := s.compileNode(branch); err != nil {
				return nil, err
			}
		}
		cd, ok := s.findBlockDef(n.Name)
		if ok && cd.parser != nil {
			renderer, err := cd.parser(*n)
			if err != nil {
				return nil, err
			}
			n.renderer = renderer
		}
	case *ASTSeq:
		for _, child := range n.Children {
			if _, err := s.compileNode(child); err != nil {
				return nil, err
			}
		}
	}
	return n, nil
}
