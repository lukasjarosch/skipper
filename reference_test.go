package skipper_test

// func TestParseValueReferencesNilReference(t *testing.T) {
// 	_, err := ParseValueReferences(nil)
// 	assert.ErrorIs(t, err, ErrReferenceSourceIsNil)
// }
//
// func TestParseValueReferencesNew(t *testing.T) {
// 	type input struct {
// 		path               data.Path
// 		targetPath         data.Path
// 		absoluteTargetPath data.Path
// 		targetValue        data.Value
// 		children           []input // only one layer deep!
// 	}
// 	tests := []struct {
// 		name     string
// 		input    []input
// 		expected []ValueReference
// 	}{
// 		{
// 			name: "asdf",
// 			input: []input{
// 				{
// 					path:               data.NewPath("abs.ref.path"),
// 					targetPath:         data.NewPath("foo"),
// 					absoluteTargetPath: data.NewPath("data.foo"),
// 					targetValue:        data.NewValue("${bar}"),
// 					children: []input{
// 						{
// 							path:               data.NewPath("abs.ref.path.test"),
// 							targetPath:         data.NewPath("bar"),
// 							absoluteTargetPath: data.NewPath("data.bar"),
// 							targetValue:        data.NewValue("ohai"),
// 							children:           nil,
// 						},
// 					},
// 				},
// 			},
// 			expected: []ValueReference{
// 				{
// 					Path:               data.NewPath("abs.ref.path"),
// 					TargetPath:         data.NewPath("foo"),
// 					AbsoluteTargetPath: data.NewPath("data.foo"),
// 				},
// 				{
// 					Path:               data.NewPath("abs.ref.path.test"),
// 					TargetPath:         data.NewPath("bar"),
// 					AbsoluteTargetPath: data.NewPath("data.bar"),
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			source := mocks.NewMockReferenceValueSource(t)
//
// 			valueMap := make(map[string]data.Value, len(tt.input))
// 			for _, in := range tt.input {
// 				valueMap[in.path.String()] = in.targetValue
// 			}
// 			spew.Dump(valueMap)
// 			source.EXPECT().Values().Return(valueMap)
//
// 			for _, input := range tt.input {
// 				source.EXPECT().AbsolutePath(input.targetPath, input.path).Return(input.absoluteTargetPath, nil)
// 				source.EXPECT().GetPath(input.absoluteTargetPath).Return(input.targetValue, nil)
//
// 				for _, child := range input.children {
// 					source.EXPECT().AbsolutePath(child.targetPath, child.path).Return(child.absoluteTargetPath, nil)
// 					source.EXPECT().GetPath(child.absoluteTargetPath).Return(child.targetValue, nil)
// 				}
// 			}
//
// 			references, err := ParseValueReferences(source)
//
// 			assert.NoError(t, err)
// 			assert.NotNil(t, references)
// 		})
// 	}
// }
//
// func TestParseValueReferences(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		values   map[string]data.Value
// 		expected []ValueReference
// 	}{
// 		{
// 			name: "nil value",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue(nil),
// 			},
// 			expected: nil,
// 		},
// 		{
// 			name: "no reference",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue("test"),
// 				"bar": data.NewValue("some text"),
// 			},
// 			expected: nil,
// 		},
// 		{
// 			// nested references are not really supported
// 			// in this case only the inner references are parsed correctly
// 			name: "nested references",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue("${foo:${bar}}"),
// 				"bar": data.NewValue("${${foo}}"),
// 				"baz": data.NewValue("${foo${bar}}"),
// 			},
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath("foo"),
// 					TargetPath: data.NewPath("bar"),
// 				},
// 				{
// 					Path:       data.NewPath("bar"),
// 					TargetPath: data.NewPath("foo"),
// 				},
// 				{
// 					Path:       data.NewPath("baz"),
// 					TargetPath: data.NewPath("bar"),
// 				},
// 			},
// 		},
// 		{
// 			name: "malformed references",
// 			values: map[string]data.Value{
// 				"empty":          data.NewValue("${}"),
// 				"dot_separator":  data.NewValue("${a.b}"),
// 				"unclosed":       data.NewValue("${a"),
// 				"illegal_chars":  data.NewValue("${a*b%}"),
// 				"missing_dollar": data.NewValue("{a:b}"),
// 			},
// 			expected: nil,
// 		},
// 		{
// 			name: "simple references",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue("${foo}"),
// 				"bar": data.NewValue("${bar}"),
// 			},
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath("foo"),
// 					TargetPath: data.NewPath("foo"),
// 				},
// 				{
// 					Path:       data.NewPath("bar"),
// 					TargetPath: data.NewPath("bar"),
// 				},
// 			},
// 		},
// 		{
// 			name: "long references",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue("${this:is:a:path}"),
// 				"bar": data.NewValue("${an:even:longer:path:why:not}"),
// 			},
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath("foo"),
// 					TargetPath: data.NewPath("this.is.a.path"),
// 				},
// 				{
// 					Path:       data.NewPath("bar"),
// 					TargetPath: data.NewPath("an.even.longer.path.why.not"),
// 				},
// 			},
// 		},
// 		{
// 			name: "embedded references",
// 			values: map[string]data.Value{
// 				"foo": data.NewValue("Hello there, ${name}"),
// 				"bar": data.NewValue("Ohai, ${name:first} ${name:last}"),
// 			},
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath("foo"),
// 					TargetPath: data.NewPath("name"),
// 				},
// 				{
// 					Path:       data.NewPath("bar"),
// 					TargetPath: data.NewPath("name.first"),
// 				},
// 				{
// 					Path:       data.NewPath("bar"),
// 					TargetPath: data.NewPath("name.last"),
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			source := mocks.NewMockReferenceValueSource(t)
// 			source.EXPECT().Values().Return(tt.values)
//
// 			for _, expect := range tt.expected {
// 				source.EXPECT().AbsolutePath(expect.TargetPath, expect.Path).Return(expect.AbsoluteTargetPath, nil)
// 			}
//
// 			references, err := ParseValueReferences(source)
// 			assert.NoError(t, err)
// 			if tt.expected == nil {
// 				assert.Nil(t, references)
// 			} else {
// 				assert.ElementsMatch(t, references, tt.expected)
// 			}
// 			source.AssertExpectations(t)
// 		})
// 	}
// }
//
// func TestReferencesInValue(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		value    data.Value
// 		expected []ValueReference
// 	}{
// 		{
// 			name:     "no reference",
// 			value:    data.NewValue("hello there"),
// 			expected: nil,
// 		},
// 		{
// 			name:  "one reference",
// 			value: data.NewValue("hello there, ${name}"),
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath(""),
// 					TargetPath: data.NewPath("name"),
// 				},
// 			},
// 		},
// 		{
// 			name:  "multiple references",
// 			value: data.NewValue("${greeting}, ${name:first} ${name:last}"),
// 			expected: []ValueReference{
// 				{
// 					Path:       data.NewPath(""),
// 					TargetPath: data.NewPath("greeting"),
// 				},
// 				{
// 					Path:       data.NewPath(""),
// 					TargetPath: data.NewPath("name.first"),
// 				},
// 				{
// 					Path:       data.NewPath(""),
// 					TargetPath: data.NewPath("name.last"),
// 				},
// 			},
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			references := ReferencesInValue(tt.value)
//
// 			if tt.expected == nil {
// 				assert.Nil(t, references)
// 			} else {
// 				assert.ElementsMatch(t, references, tt.expected)
// 			}
// 		})
// 	}
// }
//
// func TestResolveReferences(t *testing.T) {
// 	type input struct {
// 		refTargetPath string
// 		getPathErr    error
// 		value         data.Value
// 	}
//
// 	tests := []struct {
// 		name     string
// 		input    []input
// 		expected []ValueReference
// 		err      error
// 	}{
// 		{
// 			name: "simple references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("bar"),
// 				},
// 				{
// 					refTargetPath: "bar",
// 					getPathErr:    nil,
// 					value:         data.NewValue("baz"),
// 				},
// 			},
// 			expected: []ValueReference{
// 				{
// 					TargetPath: data.NewPath("bar"),
// 					Path:       data.Path{}, // path is not relevant for ResolveReferences
// 				},
// 				{
// 					TargetPath: data.NewPath("foo"),
// 					Path:       data.Path{},
// 				},
// 			},
// 		},
// 		{
// 			name: "invalid references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    ErrPathNotFound,
// 					value:         data.NilValue,
// 				},
// 			},
// 			expected: nil,
// 			err:      ErrUndefinedReferenceTarget,
// 		},
// 		{
// 			name: "duplicate references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("bar"),
// 				},
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("bar"),
// 				},
// 			},
// 			expected: []ValueReference{
// 				{
// 					TargetPath: data.NewPath("foo"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("foo"),
// 					Path:       data.Path{},
// 				},
// 			},
// 			err: nil,
// 		},
// 		{
// 			name: "self-referencing",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${foo}"),
// 				},
// 			},
// 			expected: nil,
// 			err:      ErrSelfReferencingReference,
// 		},
// 		{
// 			name: "cyclic references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${bar}"),
// 				},
// 				{
// 					refTargetPath: "bar",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${baz}"),
// 				},
// 				{
// 					refTargetPath: "baz",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${foo}"),
// 				},
// 			},
// 			expected: nil,
// 			err:      ErrCyclicReference,
// 		},
// 		{
// 			name: "simple chained references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${bar}"),
// 				},
// 				{
// 					refTargetPath: "bar",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${baz}"),
// 				},
// 				{
// 					refTargetPath: "baz",
// 					getPathErr:    nil,
// 					value:         data.NewValue("ohai"),
// 				},
// 			},
// 			expected: []ValueReference{
// 				{
// 					TargetPath: data.NewPath("baz"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("bar"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("foo"),
// 					Path:       data.Path{},
// 				},
// 			},
// 			err: nil,
// 		},
// 		{
// 			name: "complex chained references",
// 			input: []input{
// 				{
// 					refTargetPath: "foo",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${bar} ${baz}"),
// 				},
// 				{
// 					refTargetPath: "bar",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${baz}"),
// 				},
// 				{
// 					refTargetPath: "baz",
// 					getPathErr:    nil,
// 					value:         data.NewValue("ohai ${name}"),
// 				},
// 				{
// 					refTargetPath: "name",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${peter}"),
// 				},
// 				{
// 					refTargetPath: "peter",
// 					getPathErr:    nil,
// 					value:         data.NewValue("${first} ${last}"),
// 				},
// 				{
// 					refTargetPath: "first",
// 					getPathErr:    nil,
// 					value:         data.NewValue("Peter"),
// 				},
// 				{
// 					refTargetPath: "last",
// 					getPathErr:    nil,
// 					value:         data.NewValue("Parker"),
// 				},
// 			},
// 			expected: []ValueReference{
// 				{
// 					TargetPath: data.NewPath("first"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("last"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("peter"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("name"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("baz"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("bar"),
// 					Path:       data.Path{},
// 				},
// 				{
// 					TargetPath: data.NewPath("foo"),
// 					Path:       data.Path{},
// 				},
// 			},
// 			err: nil,
// 		},
// 	}
//
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			source := mocks.NewMockReferenceSourceGetter(t)
//
// 			for _, ref := range tt.input {
// 				source.EXPECT().GetPath(data.NewPath(ref.refTargetPath)).Return(ref.value, ref.getPathErr)
// 			}
//
// 			// convert input to slice of references
// 			var references []ValueReference
// 			for _, in := range tt.input {
// 				references = append(references, ValueReference{
// 					TargetPath: data.NewPath(in.refTargetPath),
// 					Path:       data.Path{}, // unused in ResolveReferences
// 				})
// 			}
//
// 			result, err := ResolveReferences(references, source)
//
// 			if tt.err != nil {
// 				assert.ErrorIs(t, err, tt.err)
// 				assert.Nil(t, result)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.NotNil(t, result)
// 				assert.Equal(t, tt.expected, result)
// 				assert.Len(t, result, len(references))
//
// 				// TODO: finish
// 			}
//
// 			source.AssertExpectations(t)
// 		})
// 	}
// }
//
// // TODO: test with ReferenceSourceRelativeGetter
//
// type ReferenceTestSuite struct {
// 	suite.Suite
// }
//
// // testCreateRegistry quickly creates a registry with files from the given rootPath
// // and ensures that the registry is properly created.
// func (suite *ReferenceTestSuite) createRegistry(rootPath string) *Registry {
// 	classFiles, err := DiscoverFiles(rootPath, codec.YamlPathSelector)
// 	assert.NoError(suite.T(), err)
//
// 	registry, err := NewRegistryFromFiles(classFiles, func(filePaths []string) ([]*Class, error) {
// 		return ClassLoader(rootPath, classFiles, codec.NewYamlCodec())
// 	})
// 	assert.NoError(suite.T(), err)
// 	assert.NotNil(suite.T(), registry)
//
// 	return registry
// }
//
// func (suite *ReferenceTestSuite) createInventory(rootPath string) *Inventory {
// 	dataPath := filepath.Join(rootPath, "data")
// 	targetsPath := filepath.Join(rootPath, "targets")
//
// 	dataRegistry := suite.createRegistry(dataPath)
// 	targetsRegistry := suite.createRegistry(targetsPath)
//
// 	inventory, err := NewInventory()
// 	assert.NoError(suite.T(), err)
//
// 	err = inventory.RegisterScope(DataScope, dataRegistry)
// 	assert.NoError(suite.T(), err)
// 	err = inventory.RegisterScope(TargetsScope, targetsRegistry)
// 	assert.NoError(suite.T(), err)
//
// 	return inventory
// }
//
// // testParse is a helper to quickly test the 'ParseReferences' function
// // which assumes that the function does not return an error.
// // It checks that all passed references are returned.
// func (suite *ReferenceTestSuite) testParse(source OldParseSource, expected []ValueReference) []ValueReference {
// 	// TEST: ensure source cannot be nil
// 	_, err := ParseReferences(nil)
// 	assert.ErrorIs(suite.T(), err, ErrReferenceSourceIsNil)
//
// 	// TEST: ensure that parsing succeeds and returns the expected references
// 	references, err := ParseReferences(source)
// 	assert.NoError(suite.T(), err)
// 	assert.NotNil(suite.T(), references)
// 	assert.Len(suite.T(), references, len(expected))
// 	for _, reference := range references {
// 		assert.Contains(suite.T(), expected, reference)
// 	}
//
// 	return references
// }
//
// // testResolve performs the default set of tests for ResolveReferences.
// // The given expected references must be in the order in which the ResolveReferences returns them.
// func (suite *ReferenceTestSuite) testResolve(source DataSetterGetter, expected []ValueReference) []ValueReference {
// 	// TEST: ResolveReferences must not accept a nil source
// 	_, err := ResolveReferences(expected, nil)
// 	assert.ErrorIs(suite.T(), err, ErrReferenceSourceIsNil)
//
// 	// TEST: references with invalid TargetPath
// 	// We're reusing the path of an expected reference because that needs to be valid still.
// 	invalidReferences := []ValueReference{
// 		{
// 			Path:       expected[0].Path,
// 			TargetPath: data.NewPath("invalid.target.path"),
// 		},
// 		{
// 			Path:       expected[0].Path,
// 			TargetPath: data.NewPath("another.invalid.target.path"),
// 		},
// 	}
// 	resolved, err := ResolveReferences(invalidReferences, source)
// 	assert.ErrorIs(suite.T(), err, ErrUndefinedReferenceTarget)
// 	assert.Nil(suite.T(), resolved)
//
// 	// TEST: ensure that the resolving works and check whether all expected references are still within the resolved ones.
// 	resolved, err = ResolveReferences(expected, source)
// 	assert.NoError(suite.T(), err)
// 	assert.Len(suite.T(), resolved, len(expected))
//
// 	// for _, res := range resolved {
// 	// 	spew.Println(res.Name())
// 	// }
//
// 	// The expected references are sorted and since we use a stable topological sort,
// 	// the result must be exactly the same.
// 	assert.Equal(suite.T(), expected, resolved, "The resolved references must match the content and order of the expected references")
//
// 	return resolved
// }
//
// func (suite *ReferenceTestSuite) testReplace(source ReferenceSource, resolved []ValueReference, expected map[string]data.Value) {
// 	// TEST: source cannot be nil
// 	err := ReplaceReferences(resolved, nil)
// 	assert.ErrorIs(suite.T(), err, ErrReferenceSourceIsNil)
//
// 	// TEST: replacing must result in the expected values
// 	err = ReplaceReferences(resolved, source)
// 	assert.NoError(suite.T(), err)
// 	for targetPath, expectedValue := range expected {
// 		value, err := source.GetPath(data.NewPath(targetPath))
// 		assert.NoError(suite.T(), err)
// 		assert.Equal(suite.T(), expectedValue, value)
// 	}
//
// 	// // TEST: there must be no more references after replacing
// 	// parsedReferences, err := ParseReferences(source)
// 	// assert.NoError(suite.T(), err)
// 	// assert.Len(suite.T(), parsedReferences, 0)
// }
//
// // --------------------------------------------------
// // Class Reference Tests
// // --------------------------------------------------
//
// // ClassReferenceTestSuite is used to test references on a single Class (aka local references)
// type ClassReferenceTestSuite struct {
// 	ReferenceTestSuite
// }
//
// func (suite *ClassReferenceTestSuite) TestSimple() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("simple.departments.engineering.manager"),
// 			TargetPath: data.NewPath("employees.0.name"),
// 		},
// 		{
// 			Path:       data.NewPath("simple.departments.analytics.manager"),
// 			TargetPath: data.NewPath("simple.employees.1.name"),
// 		},
// 		{
// 			Path:       data.NewPath("simple.departments.marketing.manager"),
// 			TargetPath: data.NewPath("simple.employees.2.name"),
// 		},
// 		{
// 			Path:       data.NewPath("simple.projects.Project_X.department"),
// 			TargetPath: data.NewPath("simple.departments.engineering.name"),
// 		},
// 	}
//
// 	class, err := NewClass("testdata/references/local/simple.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("simple"))
// 	assert.NoError(suite.T(), err)
//
// 	// TEST: ParseReferences
// 	suite.testParse(class, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := []ValueReference{
// 		expected[3],
// 		expected[1],
// 		expected[2],
// 		expected[0],
// 	}
// 	resolved := suite.testResolve(class, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"simple.departments.engineering.manager": data.NewValue("John Doe"),
// 		"simple.departments.analytics.manager":   data.NewValue("Jane Smith"),
// 		"simple.departments.marketing.manager":   data.NewValue("Michael Johnson"),
// 		"simple.projects.Project_X.department":   data.NewValue("Engineering"),
// 	}
// 	suite.testReplace(class, resolved, expectedReplaced)
// }
//
// func (suite *ClassReferenceTestSuite) TestNested() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("nested.target"),
// 			TargetPath: data.NewPath("source"),
// 		},
// 		{
// 			Path:       data.NewPath("nested.target_array"),
// 			TargetPath: data.NewPath("source_array"),
// 		},
// 		{
// 			Path:       data.NewPath("nested.target_nested_map"),
// 			TargetPath: data.NewPath("nested_map"),
// 		},
// 		{
// 			Path:       data.NewPath("nested.target_nested_mixed"),
// 			TargetPath: data.NewPath("nested_mixed"),
// 		},
// 	}
//
// 	class, err := NewClass("testdata/references/local/nested.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("nested"))
// 	assert.NoError(suite.T(), err)
//
// 	// TEST: ParseReferences
// 	suite.testParse(class, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := []ValueReference{
// 		expected[3],
// 		expected[1],
// 		expected[2],
// 		expected[0],
// 	}
// 	resolved := suite.testResolve(class, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"nested.target": data.NewValue(map[string]interface{}{
// 			"foo": "bar",
// 			"bar": "baz",
// 		},
// 		),
// 		"nested.target_array": data.NewValue([]interface{}{
// 			"foo", "bar", "baz",
// 		}),
// 		"nested.target_nested_map": data.NewValue(map[string]interface{}{
// 			"foo": map[string]interface{}{
// 				"bar": map[string]interface{}{
// 					"baz": "qux",
// 				},
// 			},
// 		}),
// 		"nested.target_nested_mixed": data.NewValue(map[string]interface{}{
// 			"foo": []interface{}{
// 				map[string]interface{}{
// 					"bar": "baz",
// 				},
// 				"test",
// 				map[string]interface{}{
// 					"foo": map[string]interface{}{
// 						"bar": "baz",
// 					},
// 				},
// 				map[string]interface{}{
// 					"array": []interface{}{
// 						"one", "two", "three",
// 					},
// 				},
// 			},
// 		}),
// 	}
//
// 	suite.testReplace(class, resolved, expectedReplaced)
// }
//
// func (suite *ClassReferenceTestSuite) TestChained() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("chained.gotcha"),
// 			TargetPath: data.NewPath("chained.john.first_name"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.name_placeholder"),
// 			TargetPath: data.NewPath("gotcha"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.first_name"),
// 			TargetPath: data.NewPath("name_placeholder"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.greeting"),
// 			TargetPath: data.NewPath("first_name"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.ohai"),
// 			TargetPath: data.NewPath("test"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.whoop"),
// 			TargetPath: data.NewPath("tesa"),
// 		},
// 		{
// 			Path:       data.NewPath("chained.longer"),
// 			TargetPath: data.NewPath("longer_ref"),
// 		},
// 	}
//
// 	class, err := NewClass("testdata/references/local/chained.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("chained"))
// 	assert.NoError(suite.T(), err)
//
// 	// TEST: ParseReferences
// 	suite.testParse(class, expected)
//
// 	// TEST: ResolveReferences
//
// 	expectedOrder := []ValueReference{
// 		expected[0],
// 		expected[1],
// 		expected[2],
// 		expected[3],
// 		expected[6],
// 		expected[5],
// 		expected[4],
// 	}
// 	resolved := suite.testResolve(class, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"chained.greeting":         data.NewValue("Hello, John"),
// 		"chained.gotcha":           data.NewValue("John"),
// 		"chained.first_name":       data.NewValue("John"),
// 		"chained.name_placeholder": data.NewValue("John"),
// 	}
// 	suite.testReplace(class, resolved, expectedReplaced)
// }
//
// func (suite *ClassReferenceTestSuite) TestCycle() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("cycle.john"),
// 			TargetPath: data.NewPath("middle"),
// 		},
// 		{
// 			Path:       data.NewPath("cycle.name"),
// 			TargetPath: data.NewPath("john"),
// 		},
// 		{
// 			Path:       data.NewPath("cycle.middle"),
// 			TargetPath: data.NewPath("name"),
// 		},
// 	}
//
// 	class, err := NewClass("testdata/references/local/cycle.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("cycle"))
// 	assert.NoError(suite.T(), err)
//
// 	// TEST: ResolveReferences cycle detection
// 	resolved, err := ResolveReferences(expected, class)
// 	assert.ErrorIs(suite.T(), err, ErrCyclicReference)
// 	assert.Len(suite.T(), resolved, 0)
// }
//
// func (suite *ClassReferenceTestSuite) TestMulti() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("multi.project.description"),
// 			TargetPath: data.NewPath("project.name"),
// 		},
// 		{
// 			Path:       data.NewPath("multi.project.description"),
// 			TargetPath: data.NewPath("multi.project.name"),
// 		},
// 		{
// 			Path:       data.NewPath("multi.project.description"),
// 			TargetPath: data.NewPath("project.name"),
// 		},
// 		{
// 			Path:       data.NewPath("multi.project.description"),
// 			TargetPath: data.NewPath("multi.project.repo"),
// 		},
// 		{
// 			Path:       data.NewPath("multi.project.repo"),
// 			TargetPath: data.NewPath("multi.common.repo_url"),
// 		},
// 	}
//
// 	class, err := NewClass("testdata/references/local/multi.yaml", codec.NewYamlCodec(), data.NewPathFromOsPath("multi"))
// 	assert.NoError(suite.T(), err)
//
// 	// TEST: ParseReferences
// 	suite.testParse(class, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := []ValueReference{
// 		expected[4],
// 		expected[1],
// 		expected[3],
// 		expected[0],
// 		expected[2],
// 	}
// 	resolved := suite.testResolve(class, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"multi.project.repo":        data.NewValue("github.com/lukasjarosch/skipper"),
// 		"multi.project.description": data.NewValue("The skipper project is very cool. Because skipper helps in working with the Infrastructure as Data concept. The project 'skipper' is hosted on github.com/lukasjarosch/skipper.\n"),
// 	}
// 	suite.testReplace(class, resolved, expectedReplaced)
// }
//
// func TestClassReferences(t *testing.T) {
// 	suite.Run(t, new(ClassReferenceTestSuite))
// }
//
// // --------------------------------------------------
// // Registry Reference Tests
// // --------------------------------------------------
//
// // RegistryReferenceTestSuite is used to test references on a Registry
// type RegistryReferenceTestSuite struct {
// 	ReferenceTestSuite
// }
//
// func (suite *RegistryReferenceTestSuite) TestSimple() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("test.name"),
// 			TargetPath: data.NewPath("person.first_name"),
// 		},
// 		{
// 			Path:       data.NewPath("test.name"),
// 			TargetPath: data.NewPath("person.last_name"),
// 		},
// 		{
// 			Path:       data.NewPath("person.age"),
// 			TargetPath: data.NewPath("test.age"),
// 		},
// 	}
//
// 	rootPath := "testdata/references/registry/simple"
// 	registry := suite.createRegistry(rootPath)
//
// 	// TEST: ParseReferences
// 	suite.testParse(registry, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := expected
// 	resolved := suite.testResolve(registry, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"person.age": data.NewValue(30),
// 		"test.name":  data.NewValue("John Doe"),
// 	}
// 	suite.testReplace(registry, resolved, expectedReplaced)
// }
//
// func (suite *RegistryReferenceTestSuite) TestNested() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("test.target"),
// 			TargetPath: data.NewPath("nested.source"),
// 		},
// 		{
// 			Path:       data.NewPath("test.target_array"),
// 			TargetPath: data.NewPath("nested.source_array"),
// 		},
// 		{
// 			Path:       data.NewPath("test.target_nested_map"),
// 			TargetPath: data.NewPath("nested.nested_map"),
// 		},
// 		{
// 			Path:       data.NewPath("test.target_nested_mixed"),
// 			TargetPath: data.NewPath("nested.nested_mixed"),
// 		},
// 	}
//
// 	rootPath := "testdata/references/registry/nested"
// 	registry := suite.createRegistry(rootPath)
//
// 	// TEST: ParseReferences
// 	suite.testParse(registry, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := []ValueReference{
// 		expected[3],
// 		expected[1],
// 		expected[2],
// 		expected[0],
// 	}
// 	resolved := suite.testResolve(registry, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"test.target": data.NewValue(map[string]interface{}{
// 			"foo": "bar",
// 			"bar": "baz",
// 		},
// 		),
// 		"test.target_array": data.NewValue([]interface{}{
// 			"foo", "bar", "baz",
// 		}),
// 		"test.target_nested_map": data.NewValue(map[string]interface{}{
// 			"foo": map[string]interface{}{
// 				"bar": map[string]interface{}{
// 					"baz": "qux",
// 				},
// 			},
// 		}),
// 		"test.target_nested_mixed": data.NewValue(map[string]interface{}{
// 			"foo": []interface{}{
// 				map[string]interface{}{
// 					"bar": "baz",
// 				},
// 				"test",
// 				map[string]interface{}{
// 					"foo": map[string]interface{}{
// 						"bar": "baz",
// 					},
// 				},
// 				map[string]interface{}{
// 					"array": []interface{}{
// 						"one", "two", "three",
// 					},
// 				},
// 			},
// 		}),
// 	}
// 	suite.testReplace(registry, resolved, expectedReplaced)
// }
//
// func (suite *RegistryReferenceTestSuite) TestChained() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("test.full_name"),
// 			TargetPath: data.NewPath("person.first_name"),
// 		},
// 		{
// 			Path:       data.NewPath("test.full_name"),
// 			TargetPath: data.NewPath("person.last_name"),
// 		},
// 		{
// 			Path:       data.NewPath("test.greet_person"),
// 			TargetPath: data.NewPath("greeting.text"),
// 		},
// 		{
// 			Path:       data.NewPath("test.greet_person"),
// 			TargetPath: data.NewPath("full_name"),
// 		},
// 	}
//
// 	rootPath := "testdata/references/registry/chained"
// 	registry := suite.createRegistry(rootPath)
//
// 	// TEST: ParseReferences
// 	suite.testParse(registry, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := expected
// 	resolved := suite.testResolve(registry, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"test.full_name":    data.NewValue("John Doe"),
// 		"test.greet_person": data.NewValue("Hello, John Doe"),
// 	}
// 	suite.testReplace(registry, resolved, expectedReplaced)
// }
//
// func (suite *RegistryReferenceTestSuite) TestCycle() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("test.foo"),
// 			TargetPath: data.NewPath("cycle.foo"),
// 		},
// 		{
// 			Path:       data.NewPath("test.baz"),
// 			TargetPath: data.NewPath("foo"),
// 		},
// 		{
// 			Path:       data.NewPath("cycle.foo"),
// 			TargetPath: data.NewPath("bar"),
// 		},
// 		{
// 			Path:       data.NewPath("cycle.bar"),
// 			TargetPath: data.NewPath("test.baz"),
// 		},
// 	}
//
// 	rootPath := "testdata/references/registry/cycle"
// 	registry := suite.createRegistry(rootPath)
//
// 	// TEST: ResolveReferences cycle detection
// 	resolved, err := ResolveReferences(expected, registry)
// 	assert.ErrorIs(suite.T(), err, ErrCyclicReference)
// 	assert.Len(suite.T(), resolved, 0)
// }
//
// func TestRegistryReferences(t *testing.T) {
// 	suite.Run(t, new(RegistryReferenceTestSuite))
// }
//
// // --------------------------------------------------
// // Inventory Reference Tests
// // --------------------------------------------------
//
// // ClassReferenceTestSuite is used to test references on a single Class (aka local references)
// type InventoryReferenceTestSuite struct {
// 	ReferenceTestSuite
// }
//
// func (suite *InventoryReferenceTestSuite) TestSimple() {
// 	expected := []ValueReference{
// 		{
// 			Path:       data.NewPath("data.test.state"),
// 			TargetPath: data.NewPath("targets.skipper.default_state"),
// 		},
// 		{
// 			Path:       data.NewPath("data.test.foo"),
// 			TargetPath: data.NewPath("bar"),
// 		},
// 		{
// 			Path:       data.NewPath("targets.skipper.greeting"),
// 			TargetPath: data.NewPath("data.test.name"),
// 		},
// 		{
// 			Path:       data.NewPath("targets.skipper.hello"),
// 			TargetPath: data.NewPath("planet"),
// 		},
// 	}
//
// 	rootPath := "testdata/references/inventory/simple"
// 	inventory := suite.createInventory(rootPath)
//
// 	// TEST: ParseReferences
// 	suite.testParse(inventory, expected)
//
// 	// TEST: ResolveReferences
// 	expectedOrder := []ValueReference{
// 		expected[0],
// 		expected[2],
// 		expected[3],
// 		expected[1],
// 	}
// 	resolved := suite.testResolve(inventory, expectedOrder)
//
// 	// TEST: ReplaceReferences
// 	expectedReplaced := map[string]data.Value{
// 		"data.test.state":          data.NewValue("hungry"),
// 		"data.test.foo":            data.NewValue("bar"),
// 		"targets.skipper.greeting": data.NewValue("Hello there, John Doe"),
// 		"targets.skipper.hello":    data.NewValue("world"),
// 	}
// 	suite.testReplace(inventory, resolved, expectedReplaced)
// }
//
// func TestInventoryReferences(t *testing.T) {
// 	suite.Run(t, new(InventoryReferenceTestSuite))
// }
