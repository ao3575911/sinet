const fs = require('node:fs');
module.exports = async ({ github, context, core }) => {
  const config = JSON.parse(fs.readFileSync('.github/community.json', 'utf8'));
  const repo = context.repo;
  for (const label of config.labels) {
    try { await github.rest.issues.getLabel({ ...repo, name: label.name }); }
    catch (error) { if (error.status !== 404) throw error; await github.rest.issues.createLabel({ ...repo, ...label }); }
  }
  for (const [number, labels] of Object.entries(config.issue_labels)) {
    await github.rest.issues.addLabels({ ...repo, issue_number: Number(number), labels });
  }
  const { data } = await github.rest.repos.get(repo);
  if (!data.has_discussions) {
    core.warning('Discussions are disabled; owner setup tracked in issue #8. Labels were synchronized.');
    await core.summary.addHeading('Community setup').addRaw('Labels applied. Discussions blocked by repository setting; see #8.').write();
    return;
  }
  const result = await github.graphql(`query($owner:String!,$name:String!){repository(owner:$owner,name:$name){id discussionCategories(first:30){nodes{id name}} discussions(first:100){nodes{id title url}} labels(first:100){nodes{id name}}}}`, { owner: repo.owner, name: repo.repo });
  const repository = result.repository;
  for (const seed of config.discussions) {
    let discussion = repository.discussions.nodes.find(d => d.title === seed.title);
    if (!discussion) {
      const category = repository.discussionCategories.nodes.find(c => c.name === seed.category) || repository.discussionCategories.nodes.find(c => c.name === 'General');
      if (!category) { core.warning(`No suitable discussion category for ${seed.title}; see #8.`); continue; }
      const created = await github.graphql(`mutation($repositoryId:ID!,$categoryId:ID!,$title:String!,$body:String!){createDiscussion(input:{repositoryId:$repositoryId,categoryId:$categoryId,title:$title,body:$body}){discussion{id title url}}}`, { repositoryId: repository.id, categoryId: category.id, title: seed.title, body: seed.body });
      discussion = created.createDiscussion.discussion;
    }
    const ids = seed.labels.map(name => repository.labels.nodes.find(l => l.name === name)?.id).filter(Boolean);
    if (ids.length) {
      await github.graphql(`mutation($id:ID!,$labels:[ID!]!){addLabelsToLabelable(input:{labelableId:$id,labelIds:$labels}){clientMutationId}}`, { id: discussion.id, labels: ids });
    }
    core.info(`Discussion available: ${discussion.url}`);
  }
  await core.summary.addHeading('Community setup').addRaw('Labels and discussion starters synchronized.').write();
};
