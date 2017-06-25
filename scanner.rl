package main

import "strconv"

%% machine lexer;

func ScanExpression(data string) ([]Token, error) {
	cs, p, pe, eof := 0, 0, len(data), len(data)
  var (ts, te, act int)
  _ = act
  tokens := make([]Token, 0)

	%%{
    action Ident {tokens = append(tokens, Token{IdentifierType, data[ts:te], nil})}
    action Number {
      n, err := strconv.ParseFloat(data[ts:te], 64)
      if err != nil {
        panic(err)
      }
      tokens = append(tokens, Token{ValueType, "", n})
    }
    action Relation {tokens = append(tokens, Token{RelationType, data[ts:te], nil})}

    ident = (alpha | '_') . (alnum | '_')* ;
    number = '-'? (digit+ ('.' digit*)?) ;

    main := |*
      ident => Ident;
      number => Number;
      ("==" | "!=" | ">" | ">" | ">=" | "<=") => Relation;
      ("and" | "or" | "contains") => Relation;
      space+;
    *|;

		write init;
		write exec;
  }%%

	return tokens, nil
}

%% write data;