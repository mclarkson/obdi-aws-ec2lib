#!/bin/bash

git checkout master
F=/var/tmp/build.sh.$$.html
cp build-docs/header.frag.html $F
markdown README.md >> $F
cat build-docs/footer.frag.html >> $F
git checkout gh-pages
mv $F index.html
echo
echo "Do 'git commit index.html' and 'git push'"
echo
