# DB Cleaner

This job removes a small subset of archived records from the database. We basically never run proper `DELETE` queries on data except in this job.

Right now it only trims up expired OAuth2 tokens, but I really created it as a canvas for whatever other data you might think to delete regularly.
