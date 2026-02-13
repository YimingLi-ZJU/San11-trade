cp /root/San11-trade/data/san11trade.db /root/San11-trade/data/san11trade.db.bak.$(date +%Y%m%d_%H%M%S)
git pull
docker compose down
docker compose build --no-cache
docker compose up -d
docker-compose run --rm backend ./server -create-admin -admin-user=pedri -admin-pass=qw147741qw