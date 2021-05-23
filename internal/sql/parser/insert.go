package parser

import (
	"github.com/genjidb/genji/internal/expr"
	"github.com/genjidb/genji/internal/query"
	"github.com/genjidb/genji/internal/sql/scanner"
	"github.com/genjidb/genji/internal/stringutil"
)

// parseInsertStatement parses an insert string and returns a Statement AST object.
// This function assumes the INSERT token has already been consumed.
func (p *Parser) parseInsertStatement() (*query.InsertStmt, error) {
	var stmt query.InsertStmt
	var err error

	// Parse "INTO".
	if err := p.parseTokens(scanner.INTO); err != nil {
		return nil, err
	}

	// Parse table name
	stmt.TableName, err = p.parseIdent()
	if err != nil {
		pErr := err.(*ParseError)
		pErr.Expected = []string{"table_name"}
		return nil, pErr
	}

	// Parse path list: (a, b, c)
	stmt.Fields, err = p.parseFieldList()
	if err != nil {
		return nil, err
	}

	// Check if VALUES or SELECT token exists.
	tok, pos, lit := p.ScanIgnoreWhitespace()
	switch tok {
	case scanner.VALUES:
		// Parse VALUES (v1, v2, v3)
		stmt.Values, err = p.parseValues(stmt.Fields)
		if err != nil {
			return nil, err
		}
	case scanner.SELECT:
		stmt.SelectStmt, err = p.parseSelectStatement()
		if err != nil {
			return nil, err
		}
	default:
		return nil, newParseError(scanner.Tokstr(tok, lit), []string{"VALUES", "SELECT"}, pos)
	}

	stmt.Returning, err = p.parseReturning()
	if err != nil {
		return nil, err
	}

	return &stmt, nil
}

// parseFieldList parses a list of fields in the form: (path, path, ...), if exists.
// If the list is empty, it returns an error.
func (p *Parser) parseFieldList() ([]string, error) {
	// Parse ( token.
	if ok, err := p.parseOptional(scanner.LPAREN); !ok || err != nil {
		p.Unscan()
		return nil, err
	}

	// Parse path list.
	var fields []string
	var err error
	if fields, err = p.parseIdentList(); err != nil {
		return nil, err
	}

	// Parse required ) token.
	if err := p.parseTokens(scanner.RPAREN); err != nil {
		return nil, err
	}

	return fields, nil
}

// parseValues parses the "VALUES" clause of the query, if it exists.
func (p *Parser) parseValues(fields []string) ([]expr.Expr, error) {
	if len(fields) > 0 {
		return p.parseDocumentsWithFields(fields)
	}

	return p.parseLiteralDocOrParamList()
}

// parseExprListValues parses the "VALUES" clause of the query, if it exists.
func (p *Parser) parseDocumentsWithFields(fields []string) ([]expr.Expr, error) {
	var docs []expr.Expr

	// Parse first (required) value list.
	doc, err := p.parseExprListWithFields(fields)
	if err != nil {
		return nil, err
	}

	docs = append(docs, doc)

	// Parse remaining (optional) values.
	for {
		if tok, _, _ := p.ScanIgnoreWhitespace(); tok != scanner.COMMA {
			p.Unscan()
			break
		}

		doc, err := p.parseExprListWithFields(fields)
		if err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// parseParamOrDocument parses either a parameter or a document.
func (p *Parser) parseExprListWithFields(fields []string) (*expr.KVPairs, error) {
	list, err := p.parseExprList(scanner.LPAREN, scanner.RPAREN)
	if err != nil {
		return nil, err
	}

	var pairs expr.KVPairs
	pairs.Pairs = make([]expr.KVPair, len(list))

	if len(fields) != len(list) {
		return nil, stringutil.Errorf("%d values for %d fields", len(list), len(fields))
	}

	for i := range list {
		pairs.Pairs[i].K = fields[i]
		pairs.Pairs[i].V = list[i]
	}

	return &pairs, nil
}

// parseExprListValues parses the "VALUES" clause of the query, if it exists.
func (p *Parser) parseLiteralDocOrParamList() ([]expr.Expr, error) {
	var docs []expr.Expr

	// Parse first (required) value list.
	doc, err := p.parseParamOrDocument()
	if err != nil {
		return nil, err
	}

	docs = append(docs, doc)

	// Parse remaining (optional) values.
	for {
		if tok, _, _ := p.ScanIgnoreWhitespace(); tok != scanner.COMMA {
			p.Unscan()
			break
		}

		doc, err := p.parseParamOrDocument()
		if err != nil {
			return nil, err
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

// parseParamOrDocument parses either a parameter or a document.
func (p *Parser) parseParamOrDocument() (expr.Expr, error) {
	// Parse a param first
	prm, err := p.parseParam()
	if err != nil {
		return nil, err
	}
	if prm != nil {
		return prm, nil
	}

	// If not a param, start over
	p.Unscan()

	// Expect a document
	return p.ParseDocument()
}

func (p *Parser) parseReturning() ([]expr.Expr, error) {
	// Parse RETURNING clause: RETURNING expr [AS alias]
	if ok, err := p.parseOptional(scanner.RETURNING); !ok || err != nil {
		return nil, err
	}

	return p.parseProjectedExprs()
}
