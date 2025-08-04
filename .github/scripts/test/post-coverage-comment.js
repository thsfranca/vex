const fs = require('fs');

module.exports = async ({github, context}) => {
  // Read the coverage report
  let coverageReport = '';
  try {
    coverageReport = fs.readFileSync('coverage-report.md', 'utf8');
  } catch (error) {
    coverageReport = '## ⚠️ Coverage Report Not Available\n\nCoverage analysis could not be completed for this PR.';
  }
  
  // Find existing coverage comment
  const { data: comments } = await github.rest.issues.listComments({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: context.issue.number,
  });
  
  const existingComment = comments.find(comment => 
    comment.user.login === 'github-actions[bot]' && 
    comment.body.includes('Test Coverage Report')
  );
  
  // Create or update comment
  if (existingComment) {
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: existingComment.id,
      body: coverageReport
    });
    console.log('Updated existing coverage comment');
  } else {
    await github.rest.issues.createComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: context.issue.number,
      body: coverageReport
    });
    console.log('Created new coverage comment');
  }
};