# cov-diff
A github action that shows coverage on new code for a PR

This is action is basically just a wrapper around the cov-diff utility. All it does is it gets a file containing the diff between two commits,
the coverage file (coverprofile) and the path to the code and it then computes the percentage of the new code that is covered by tests. I usually pair this 
with another action that leaves a comment on the PR for visibility.

# How does it work?
The algorithm computes 3 sets of intervals, where interval is a struct that has a Start and an End value (range). 
- The first set contains all the intervals of the lines that changed. For example `[{Start: 0, End: 10}, {Start: 15, End: 20}]` means that lines 
0-10 and 15-20 (all inclusive) have changed
- The second set contains all the intervals of the lines we care about. These are the intervals where functions exist (includes function definition and body).
We get this using the Go parser from the standard library so this only works for Go. These are the interesting lines and the ones we want to cover with tests
anything outside of functions such as struct definitions and var declerations are ignored. We also ignore anything under vendor and everything in `package main`
- The third set contains the intervals of all the lines covered by tests.

Finally the total lines that need to be covered are: `Union(diffLines, functionLines)` - The lines that changed and are inside functions
And the covered lines are: `Union(coveredLines, Union(diffLines, functionLines))` - The lines that changed, are in functions, and are covered

Here's an example github action setup (you need to set `fetch-depth: 0` for the checkout action otherwise it won't find the branches to generate the diff)
```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
        with:
          fetch-depth: 0

      - name: run tests
        run: |
          go test ./... -coverprofile=coverage.out
          
      - name: generate diff
        run: |
          git diff origin/${{ github.base_ref }} origin/${{ github.head_ref }} > pr.diff
          
      - name: compute new code coverage
        id: covdiffaction
        uses: panagiotisptr/cov-diff@main
        with:
          path: .
          coverprofile: coverage.out
          diff: pr.diff
          module: github.com/panagiotisptr/cov-diff

      - name: Comment
        uses: mshick/add-pr-comment@v2
        with:
          message: |
            Coverage on new code: ${{ steps.covdiffaction.outputs.covdiff }}%
```
