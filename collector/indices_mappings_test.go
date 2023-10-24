// Copyright 2021 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
)

func TestMapping(t *testing.T) {
	// Testcases created using:
	//  docker run -p 9200:9200 -e "discovery.type=single-node" elasticsearch:7.8.0
	//  curl -XPUT http://localhost:9200/twitter
	//  curl -XPUT http://localhost:9200/facebook
	/*  curl -XPUT http://localhost:9200/twitter/_mapping -H 'Content-Type: application/json' -d'{
	    "properties": {
	        "email": {
	            "type": "keyword"
	        },
	        "phone": {
	            "type": "keyword"
	        }
	    }
	}'*/
	/*  curl -XPUT http://localhost:9200/facebook/_mapping -H 'Content-Type: application/json' -d'{
	    "properties": {
	        "name": {
	            "type": "text",
	            "fields": {
	                "raw": {
	                    "type": "keyword"
	                }
	            }
	        },
	        "contact": {
	            "properties": {
	                "email": {
	                    "type": "text",
	                    "fields": {
	                        "raw": {
	                            "type": "keyword"
	                        }
	                    }
	                },
	                "phone": {
	                    "type": "text"
	                }
	            }
	        }
	    }
	}'*/
	//  curl http://localhost:9200/_all/_mapping
	tcs := map[string]string{
		"7.8.0": `{
			"facebook": {
			  "mappings": {
				"properties": {
				  "contact": {
					"properties": {
					  "email": {
						"type": "text",
						"fields": {
						  "raw": {
							"type": "keyword"
						  }
						}
					  },
					  "phone": {
						"type": "text"
					  }
					}
				  },
				  "name": {
					"type": "text",
					"fields": {
					  "raw": {
						"type": "keyword"
					  }
					}
				  }
				}
			  }
			},
			"twitter": {
			  "mappings": {
				"properties": {
				  "email": {
					"type": "keyword"
				  },
				  "phone": {
					"type": "keyword"
				  }
				}
			  }
			}
		  }`,
	}
	for ver, out := range tcs {
		for hn, handler := range map[string]http.Handler{
			"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, out)
			}),
		} {
			ts := httptest.NewServer(handler)
			defer ts.Close()

			u, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("Failed to parse URL: %s", err)
			}
			c := NewIndicesMappings(log.NewNopLogger(), http.DefaultClient, u)
			imr, err := c.fetchAndDecodeIndicesMappings()
			if err != nil {
				t.Fatalf("Failed to fetch or decode indices mappings: %s", err)
			}
			t.Logf("[%s/%s] All Indices Mappings Response: %+v", hn, ver, imr)

			response := *imr
			if *response["facebook"].Mappings.Properties["contact"].Properties["phone"].Type != "text" {
				t.Errorf("Marshalling error at facebook.contact.phone")
			}

			if *response["facebook"].Mappings.Properties["contact"].Properties["email"].Fields["raw"].Type != "keyword" {
				t.Errorf("Marshalling error at facebook.contact.email.raw")
			}

			if *response["facebook"].Mappings.Properties["name"].Type != "text" {
				t.Errorf("Marshalling error at facebook.name")
			}

			if *response["facebook"].Mappings.Properties["name"].Fields["raw"].Type != "keyword" {
				t.Errorf("Marshalling error at facebook.name.raw")
			}

			if *response["twitter"].Mappings.Properties["email"].Type != "keyword" {
				t.Errorf("Marshalling error at twitter.email")
			}

			if *response["twitter"].Mappings.Properties["phone"].Type != "keyword" {
				t.Errorf("Marshalling error at twitter.phone")
			}

		}
	}
}

func TestIndexMappingFieldCount(t *testing.T) {

	testIndexNumFields := 40.0
	testIndexName := "test-data-2023.01.20"

	rawMapping := `{
			"test-data-2023.01.20": {
				"mappings": {
					"properties": {
						"data": {
							"type": "object",
							"properties": {
								"field1": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field10": {
									"type": "long"
								},
								"field2": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field3": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field4": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field5": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field6": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field7": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field8": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field9": {
									"type": "long"
								}
							}
						},
						"data2": {
							"properties": {
								"field1": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field2": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field3": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field4": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"field5": {
									"type": "text",
									"fields": {
										"keyword": {
											"type": "keyword",
											"ignore_above": 256
										}
									}
								},
								"nested_field6": {
									"properties": {
										"field1": {
											"type": "text",
											"fields": {
												"keyword": {
													"type": "keyword",
													"ignore_above": 256
												}
											}
										},
										"field2": {
											"type": "text",
											"fields": {
												"keyword": {
													"type": "keyword",
													"ignore_above": 256
												}
											}
										},
										"field3": {
											"type": "text",
											"fields": {
												"keyword": {
													"type": "keyword",
													"ignore_above": 256
												}
											}
										},
										"field4": {
											"type": "text",
											"fields": {
												"keyword": {
													"type": "keyword",
													"ignore_above": 256
												}
											}
										},
										"field5": {
											"type": "long"
										}
									}
								}
							}
						}
					}
				}
			}
		}`

	for _, handler := range map[string]http.Handler{
		"plain": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, rawMapping)
		}),
	} {

		ts := httptest.NewServer(handler)
		defer ts.Close()

		u, err := url.Parse(ts.URL)
		if err != nil {
			t.Fatalf("Failed to parse URL: %s", err)
		}
		c := NewIndicesMappings(log.NewNopLogger(), http.DefaultClient, u)
		indicesMappingsResponse, err := c.fetchAndDecodeIndicesMappings()
		if err != nil {
			t.Fatalf("Failed to fetch or decode indices mappings: %s", err)
		}

		response := *indicesMappingsResponse
		mapping := response[testIndexName]
		totalFields := countFieldsRecursive(mapping.Mappings.Properties, 0)
		if totalFields != testIndexNumFields {
			t.Errorf("Number of actual fields in index doesn't match the count returned by the recursive countFieldsRecursive function")
		}

	}

}
