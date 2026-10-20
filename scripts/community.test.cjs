const { test } = require('node:test');
const assert = require('node:assert/strict');
const synchronize = require('./community.cjs');
const config = require('../.github/community.json');
function fixture(enabled) {
  const labels = new Map(); const discussions = []; const warnings = []; let created = 0;
  const summary = { addHeading(){return this}, addRaw(){return this}, async write(){} };
  const core = { warning(message){warnings.push(message)}, info(){}, summary };
  const github = {
    rest: {
      issues: {
        async getLabel({name}){if(!labels.has(name)){const e=new Error('not found');e.status=404;throw e}return {data:labels.get(name)}},
        async createLabel(label){assert(!labels.has(label.name));labels.set(label.name,{id:'label-'+label.name,...label})},
        async addLabels({issue_number,labels:values}){assert(issue_number>=1);assert(values.every(name=>labels.has(name)))}
      },
      repos: { async get(){return {data:{has_discussions:enabled}}} }
    },
    async graphql(query,variables){
      if(query.startsWith('query')){assert.equal(variables.name,'sinet');return {repository:{id:'repo',discussionCategories:{nodes:[{id:'general',name:'General'},{id:'ideas',name:'Ideas'}]},discussions:{nodes:discussions},labels:{nodes:[...labels.values()]}}}}
      if(query.includes('createDiscussion')){created++;const discussion={id:'discussion-'+created,title:variables.title,url:'https://example.invalid/'+created};discussions.push(discussion);return {createDiscussion:{discussion}}}
      assert(query.includes('addLabelsToLabelable'));assert(variables.labels.length>0);return {};
    }
  };
  return { github, context:{repo:{owner:'ao3575911',repo:'sinet'}}, core, labels, discussions, warnings, created:()=>created };
}
test('labels synchronize idempotently and disabled Discussions is reported', async()=>{
 const f=fixture(false);await synchronize(f);await synchronize(f);
 assert.equal(f.labels.size,config.labels.length);assert.equal(f.created(),0);assert.equal(f.warnings.length,2);
});
test('discussion starters are created once and labelled',async()=>{
 const f=fixture(true);await synchronize(f);await synchronize(f);
 assert.equal(f.created(),3);assert.equal(f.discussions.length,3);assert.equal(f.warnings.length,0);
});
