#!/bin/bash

git checkout master
F=/var/tmp/build.sh.$$.html
cp header.frag.html $F
markdown README.md >> $F
cat footer.frag.html >> $F
git checkout gh-pages
mv $F index.html
echo
echo "Do 'git commit index.html' and 'git push'"
echo
