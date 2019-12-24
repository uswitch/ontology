# coding: utf-8
require 'rspec'

require_relative '../../../lib/ontology/source/docker'

# stolen from github.com/distribution/reference
RSpec.describe "#parse_image_reference" do
  it "should parse references properly" do


    tests = [
      {
		input:      "test_com",
		repository: "test_com",
	  },
	  {
		input:      "test.com:tag",
		repository: "test.com",
		tag:        "tag",
	  },
	  {
		input:      "test.com:5000",
		repository: "test.com",
		tag:        "5000",
	  },
	  {
		input:      "test.com/repo:tag",
		domain:     "test.com",
		repository: "test.com/repo",
		tag:        "tag",
	  },
	  {
		input:      "test:5000/repo",
		domain:     "test:5000",
		repository: "test:5000/repo",
	  },
	  {
		input:      "test:5000/repo:tag",
		domain:     "test:5000",
		repository: "test:5000/repo",
		tag:        "tag",
	  },
	  {
		input:      "test:5000/repo@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		domain:     "test:5000",
		repository: "test:5000/repo",
		digest:     "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	  },
	  {
		input:      "test:5000/repo:tag@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		domain:     "test:5000",
		repository: "test:5000/repo",
		tag:        "tag",
		digest:     "sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	  },
	  {
		input:      "test:5000/repo",
		domain:     "test:5000",
		repository: "test:5000/repo",
	  },
	  {
		input: "",
		err:   "ErrNameEmpty",
	  },
	  {
		input: ":justtag",
		err:   "ErrReferenceInvalidFormat",
	  },
	  {
		input: "@sha256:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		err:   "ErrReferenceInvalidFormat",
	  },
	  {
		input: "repo@sha256:ffffffffffffffffffffffffffffffffff",
		err:   "digest.ErrDigestInvalidLength",
	  },
	  {
		input: "validname@invaliddigest:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
		err:   "digest.ErrDigestUnsupported",
	  },
	  {
		input: "Uppercase:tag",
		err:   "ErrReferenceInvalidFormat",
	  },
	  {
		input: "test:5000/Uppercase/lowercase:tag",
		err:   "ErrReferenceInvalidFormat",
	  },
	  {
		input:      "lowercase:Uppercase",
		repository: "lowercase",
		tag:        "Uppercase",
	  },
	  {
		input: ("a/" * 128) + "a:tag",
		err:   "ErrNameTooLong",
	  },
	  {
		input:      ("a/" * 127) + "a:tag-puts-this-over-max",
		domain:     "a",
		repository: ("a/" * 127) + "a",
		tag:        "tag-puts-this-over-max",
	  },
	  {
		input: "aa/asdf$$^/aa",
		err:   "ErrReferenceInvalidFormat",
	  },
	  {
		input:      "sub-dom1.foo.com/bar/baz/quux",
		domain:     "sub-dom1.foo.com",
		repository: "sub-dom1.foo.com/bar/baz/quux",
	  },
	  {
		input:      "sub-dom1.foo.com/bar/baz/quux:some-long-tag",
		domain:     "sub-dom1.foo.com",
		repository: "sub-dom1.foo.com/bar/baz/quux",
		tag:        "some-long-tag",
	  },
	  {
		input:      "b.gcr.io/test.example.com/my-app:test.example.com",
		domain:     "b.gcr.io",
		repository: "b.gcr.io/test.example.com/my-app",
		tag:        "test.example.com",
	  },
	  {
		input:      "xn--n3h.com/myimage:xn--n3h.com", # ☃.com in punycode
		domain:     "xn--n3h.com",
		repository: "xn--n3h.com/myimage",
		tag:        "xn--n3h.com",
	  },
	  {
		input:      "xn--7o8h.com/myimage:xn--7o8h.com@sha512:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", # 🐳.com in punycode
		domain:     "xn--7o8h.com",
		repository: "xn--7o8h.com/myimage",
		tag:        "xn--7o8h.com",
		digest:     "sha512:ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
	  },
	  {
		input:      "foo_bar.com:8080",
		repository: "foo_bar.com",
		tag:        "8080",
	  },
	  {
		input:      "foo/foo_bar.com:8080",
		domain:     "foo",
		repository: "foo/foo_bar.com",
		tag:        "8080",
	  },

    ]

    tests.each { |test|
      if test[:err]
        expect {
          parse_image_reference(test[:input])
        }.to raise_error(test[:err])
      else
        out = parse_image_reference(test[:input])

        expected = test.clone
        expected.delete(:input)

        expect(out).to match(hash_including(**expected))
      end
    }
  end
end
