#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

module.exports = async ({ github, context }) => {
  console.log('Starting coverage comment process...');
  console.log('Event name:', context.eventName);
  console.log('Repository:', context.repo);
  console.log('PR number:', context.issue?.number);

  // Check if this is a PR
  if (context.eventName !== 'pull_request') {
    console.log('Not a pull request, skipping coverage comment');
    return;
  }

  const prNumber = context.issue.number;
  if (!prNumber) {
    console.log('No PR number found, skipping coverage comment');
    return;
  }

  // Read the coverage report file
  const reportPath = path.join(process.cwd(), 'coverage-report.md');
  console.log('Looking for coverage report at:', reportPath);

  if (!fs.existsSync(reportPath)) {
    console.log('Coverage report file not found, skipping comment');
    return;
  }

  let coverageContent;
  try {
    coverageContent = fs.readFileSync(reportPath, 'utf8');
    console.log('Coverage report read successfully, length:', coverageContent.length);
  } catch (error) {
    console.error('Error reading coverage report:', error.message);
    return;
  }

  if (!coverageContent || coverageContent.trim().length === 0) {
    console.log('Coverage report is empty, skipping comment');
    return;
  }

  try {
    // Get existing comments to see if we already posted one
    console.log('Fetching existing PR comments...');
    const { data: comments } = await github.rest.issues.listComments({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: prNumber,
    });

    console.log(`Found ${comments.length} existing comments`);

    // Look for existing coverage comment
    const coverageCommentMarker = '## 📊 Test Coverage Report';
    const existingComment = comments.find(comment => 
      comment.body.includes(coverageCommentMarker)
    );

    if (existingComment) {
      console.log('Found existing coverage comment, updating...');
      await github.rest.issues.updateComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        comment_id: existingComment.id,
        body: coverageContent,
      });
      console.log('Coverage comment updated successfully');
    } else {
      console.log('Creating new coverage comment...');
      await github.rest.issues.createComment({
        owner: context.repo.owner,
        repo: context.repo.repo,
        issue_number: prNumber,
        body: coverageContent,
      });
      console.log('Coverage comment created successfully');
    }
  } catch (error) {
    console.error('Error posting coverage comment:', error.message);
    console.error('Error details:', error);
    throw error;
  }
};