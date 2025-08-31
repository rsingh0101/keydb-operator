gcloud container clusters create my-operator-dev \
  --num-nodes=1 \
  --machine-type=e2-small \
  --preemptible \
  --region=asia-south1 \
  --disk-size=20 \
  --enable-autoupgrade \
  --enable-autorepair

