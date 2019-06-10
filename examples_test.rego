package examples

test_rule_based_noun {
  has_read_access_to[["jane", "energy-wiki"]]
}
test_rule_based_noun_fail {
  not has_read_access_to[["robert", "energy-wiki"]]
}

test_defined_classification {
  is_classified_as[["declassified-logs", "operational"]]
  count([ x | some x; is_classified_as[["declassified-logs", x]]]) == 1
}

test_inheritted_classification {
  is_classified_as[["cloud/kibana-34gg5", "restricted"]]
}
