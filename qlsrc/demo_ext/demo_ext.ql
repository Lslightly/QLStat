external predicate foo(string bar, string baz);

from string a, string b
where foo(a, b)
select a, b
