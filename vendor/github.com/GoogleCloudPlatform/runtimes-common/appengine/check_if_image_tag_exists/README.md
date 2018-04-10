#Tag Check Intermediate Cloudbuild Step
This code serves as an intermediate container in a `gcloud container builds
create` command to check for existence of the specified tag on the target
container, and exit the build if it exists. The motivation for this is to
prevent users from unintentionally overwriting a tag in a remote repository.

##Steps to Build Container Image
1. Ensure that the "Container Analysis" API is enabled in your project.
2. Ensure you're authenticated with Viewer permissions to your project's repository.
3. In your target project's cloudbuild.yaml file, add the following build step *before* your build is executed:
```
   - name: gcr.io/gcp-runtimes/check_if_tag_exists:latest
     args:
       - '--image=<target_image_path>'```
