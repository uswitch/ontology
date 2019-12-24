# coding: utf-8
require 'rspec'

require_relative '../../../lib/ontology/source/docker'

# stolen from github.com/distribution/reference
RSpec.describe "#labels_to_relations" do
  it "handle nested JSON" do

    rels = labels_to_relations(
      "/wibble", "now", {
        "cloud.rvu.ontology/relation/v1/is_part_of": "[\"asdf\",\"sdfg\"]"
      },
    )

    expect(rels.count).to eq(2)
    expect(rels[0][:properties][:b]).to eq("asdf")
    expect(rels[1][:properties][:b]).to eq("sdfg")
  end

  it "does what is should" do
    rels = labels_to_relations(
      "/wibble", "now", {
        "cloud.rvu.ontology/relation/v1/is_part_of": "asdf"
      },
    )

    expect(rels.count).to eq(1)
    expect(rels[0][:metadata][:type]).to eq("/relation/v1/is_part_of")
    expect(rels[0][:metadata][:updated_at]).to eq("now")
    expect(rels[0][:properties][:a]).to eq("/wibble")
    expect(rels[0][:properties][:b]).to eq("asdf")
  end
end
