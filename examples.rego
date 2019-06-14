 package examples

is_a_person[person] { data.people[_] = person }
is_a_team[team] { data.teams[_] = team }
is_a_role[role] { data.roles[_] = role }
is_a_repo[repo] { data.repos[_] = repo }
is_a_service[service] { data.services[_] = service }
is_a_data_classification[data_classification] { data.data_classifications[_] = data_classification }
is_a_kubernetes_pod[kubernetes_pod] { data.kubernetes_pods[_] = kubernetes_pod }
is_a_aws_s3_bucket[aws_s3_bucket] { data.aws_s3_buckets[_] = aws_s3_bucket }

is_a_asset[asset] {
  data.assets[_] = asset
} {
  is_a_service[asset]
} {
  is_a_kubernetes_pod[asset]
} {
  is_a_aws_s3_bucket[asset]
}

works_as_a[[person, role]] {
  role := data.works_as_a[person][_]
  is_a_person[person]
  is_a_role[role]
}

works_on[[person, team]] { _works_on[[person, team]] }
works_on[[person, "energy-engineers"]] {
  _works_on[[person, "energy"]]
  works_as_a[[person, "engineer"]]

  is_a_person[person]
}

_works_on[[person, team]] {
  team := data.works_on[person][_]

  is_a_person[person]
  is_a_team[team]
} {
  some i, j

  leaf_teams := data.works_on[person]
  is_a_sub_team_of[[i, j]]
  leaf_teams[_]= i
  team := j

  is_a_person[person]
  is_a_team[team]
} {
  some i, j, k

  leaf_teams := data.works_on[person]
  is_a_sub_team_of[[i, j]]
  leaf_teams[_]= i
  is_a_sub_team_of[[j, k]]
  team := k

  is_a_person[person]
  is_a_team[team]
}

is_a_sub_team_of[[team, parent_team]] {
  parent_team := data.is_a_sub_team_of[team]

  is_a_team[team]
  is_a_team[parent_team]
}

has_read_access_to[[person, repo]] {
  works_on[[person, "energy-engineers"]]
  repo := "energy-wiki"

  is_a_person[person]
  is_a_repo[repo]
}

is_classified_as[[asset, classification]] {
  classification := data.is_classified_as[asset]
  is_a_asset[asset]
  is_a_data_classification[classification]
} {
  is_a_asset[asset]
  not data.is_classified_as[asset]

  parent_classifications := [ c | some parent; is_part_of[[asset, parent]]; c := data.is_classified_as[parent] ]
  count(parent_classifications) > 0
  classification = parent_classifications[0]

  is_a_data_classification[classification]
}

is_part_of[[child,parent]] { _is_part_of[[child, parent]] }

_is_part_of[[child, parent]] {
  parent := data.is_part_of[child]

  is_a_asset[child]
  is_a_asset[parent]
} {
  some i
  data.is_part_of[i] = parent
  data.is_part_of[child] = i

  is_a_asset[child]
  is_a_asset[parent]
} {
  some i, j
  data.is_part_of[i] = parent
  data.is_part_of[j] = i
  data.is_part_of[child] = j

  is_a_asset[child]
  is_a_asset[parent]
} {
  some i, j, k
  data.is_part_of[i] = parent
  data.is_part_of[j] = i
  data.is_part_of[k] = j
  data.is_part_of[child] = k

  is_a_asset[child]
  is_a_asset[parent]
}
