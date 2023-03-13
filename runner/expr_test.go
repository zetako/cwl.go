package runner_test

import (
  "regexp"
  "testing"
)

// simple symbol
// symbol::	 {Unicode alphanumeric}+
// true: abc , 汉字，
// false:  bad_dog
func Test_expr(t *testing.T) {
  var rx = regexp.MustCompile(`\$\([\pL\pN]+(\.[\pL\pN]+)*\)`)
  epass := []string{ "$(abc你10.DEF我.mn他2)", "$(abc.dbc.mn)" }
  for _, e := range epass{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 1 {
      t.Fatal(matches)
    }
  }
  eFail := []string{ "$(bad_dog)" , "$(a.bad_dog)" }
  for _, e := range eFail{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 0 {
      t.Fatal(matches)
    }
  }
}


// simple symbol + singleq
func Test_expr2(t *testing.T) {
  var rx = regexp.MustCompile(`\$\([\pL\pN]+(\.[\pL\pN]+|\['[^'| ]+'\]|\["[^"| ]+"\])*\)`)
  epass := []string{ "$(abc你10.DEF我.mn他2)", "$(abc.dbc.mn)" ,"$(abc.dbc['abc'].mn)" ,"$(abc.dbc['我们{}-_'].mn)",
    `$(abc.dbc['abc'].mn["abc"])`, `$(abc.dbc['我们{}-_'].mn["我们{}-_"])`}
  for _, e := range epass{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 1 {
      t.Fatal(e, matches)
    }
  }
  eFail := []string{ "$(bad_dog)" , "$(a.bad_dog)","$(abc.dbc['abc''].mn)","$(ab['ab|'])","$(ab['ab cd'])"  }
  for _, e := range eFail{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 0 {
      t.Fatal(e, matches)
    }
  }
}


// simple symbol + singleq
// + doubleq + index
func Test_expr3(t *testing.T) {
  var rx = regexp.MustCompile(`\$\([\pL\pN]+(\.[\pL\pN]+|\['[^'| ]+'\]|\["[^"| ]+"\]|\[\d+\])*\)`)
  epass := []string{ "$(abc你10.DEF我.mn他2)", "$(abc.dbc.mn)" ,"$(abc.dbc['abc'].mn)" ,"$(abc.dbc['我们{}-_'].mn)",
    `$(abc.dbc['abc'].mn["abc"])`, `$(abc.dbc['我们{}-_'].mn["我们{}-_"])`,
    "$(abc.dbc.mn[0])" }
  for _, e := range epass{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 1 {
      t.Fatal(e, matches)
    }
  }
  eFail := []string{ "$(bad_dog)" , "$(a.bad_dog)","$(abc.dbc['abc''].mn)","$(ab['ab|'])","$(ab['ab cd'])" ,"$(abc.dbc.mn[-0])"  }
  for _, e := range eFail{
    matches := rx.FindAllString(e, -1)
    if len(matches) != 0 {
      t.Fatal(e, matches)
    }
  }
}

// Literals
// StringLiteral
// NumericLiteral
func Test_expr4(t *testing.T) {
  //var rx = regexp.MustCompile(`\$\((?U)(([^)(]*|())*)\)`)
  e := `$("a ")$("string")`
  matches := BetterSubmatch(e)
  t.Logf("%#v",matches)
  if len(matches) != 2 {
    t.Fatal(e, matches)
  }
  e2 := `$( 1+ (2+3))`
  matches2 := BetterSubmatch(e2)
  t.Logf("%#v",matches2)
 
}

// 匹配 合适的括号数目
// If 'Index' is present, matches and submatches are identified by byte index
// pairs within the input string: result[2*n:2*n+1] identifies the indexes of
// the nth submatch. The pair for n==0 identifies the match of the entire
// expression. If 'Index' is not present, the match is identified by the text
// of the match/submatch. If an index is negative or text is nil, it means that
// subexpression did not match any string in the input. For 'String' versions
// an empty string means either no match or an empty match.
func BetterSubmatch(in string) [][3]string  {
  var pass, l = 0 , len(in)
  var str , exp , funCode string
  var inExp , inFun bool
  var deep int
  ret := make([][3]string,0)
  for i, c := range in {
    if pass > 0 {
      pass --
      continue
    }
    // scan 3 byte
    if !inExp && !inFun {
      if l - i < 2 {
        // do nothing
      } else if c == '\\' {
        if in[i:i+2] == "\\$(" {
          pass = 2
          str += "$("
          continue
        }
        if in[i:i+2] == "\\${" {
          pass = 2
          str += "${"
          continue
        }
        if in[i:i+2] == `\\` {
          pass = 1
          str += "\\"
          continue
        }
      } else if c == '$' {
        if in[i+1] == '(' {
          pass = 1
          inExp = true
          continue
        } else if in[i+1] == '{' {
          pass = 1
          inFun = true
          continue
        }
      }
    }
    if inExp {
      if c == '(' {
        deep ++
      } else if c == ')' {
        if deep == 0 {
          // close Exp
          ret = append(ret, [3]string{str,exp, funCode})
          str, exp, funCode = "", "", ""
          inExp = false
          continue
        }
        deep--
      }
      exp = exp + in[i:i+1]
    } else if inFun {
      if c == '{' {
        deep ++
      } else if c == '}' {
        if deep == 0 {
          // close Exp
          ret = append(ret, [3]string{str,exp, funCode})
          str, exp, funCode = "", "", ""
          inFun = false
          continue
        }
        deep--
      }
      funCode = funCode + in[i:i+1]
    } else {
      str += in[i:i+1]
    }
  }
  if str != "" {
    ret = append(ret, [3]string{str,exp,funCode})
  }
  return ret
}
