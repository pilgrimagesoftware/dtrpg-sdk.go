# docs/site

Jekyll + GitHub Pages documentation site, same pattern as `dtrpg-app/rust/docs/site`.

- Versioned docs live under a directory per version (currently only `latest/`), listed in `_data/versions.yml`.
- The root `index.html` redirects to `/latest/`.
- `_layouts/default.html` renders the header, version selector, and sidebar nav from `_data/versions.yml`.
- Run `bundle install && bundle exec jekyll build` from this directory before merging changes; see `PUBLISHING.md`.
