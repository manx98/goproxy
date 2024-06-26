#!/bin/sh
echo "Change to src directory."
cd src || {
  echo "Error: Could not change to src directory"
  exit 1
}
echo "Install pnpm ..."
npm install pnpm || {
  echo "Error: Could not install pnpm"
  exit 1
}
echo "Install dependencies ..."
pnpm install || {
  echo "Error: Could not install dependencies"
  exit 1
}
export DIST_OUT_DIR=../dist || {
  echo "Error: Could not export DIST_OUT_DIR"
  exit 1
}
echo "Building ..."
pnpm build --emptyOutDir || {
  echo "Error: Could not build"
  exit 1
}
echo "Build succeeded"