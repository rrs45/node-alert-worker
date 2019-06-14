@Library("jenkins-pipeline-library") _

pipeline {
    agent { label 'skynet' }
    options {
        timeout(time: 1, unit: 'HOURS')
    }
    environment {
        SKYNET_APP = 'node-alert-worker'
    }
    parameters {
        string(name: "BUILD_NUMBER", defaultValue: "", description: "Replay build value")
    }
    stages {
        stage('Build') {
            when { branch 'master'  }
            steps {
                githubCheck(
                    'Build Image': {
                        buildImage()
                        echo "Just built image with id ${builtImage.imageId}"
                    }
                )
            }
        }

    }
        post {
        always {
            archiveBuildInfo()
        }
    }
}