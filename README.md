# RoadRunner Redis Lock
This is a proof of concept repository.
Based on the [lock plugin](https://github.com/roadrunner-server/lock) from [roadrunner-server](https://github.com/roadrunner-server)

## Configuration
```yml
redislock:
    # When polling to get a lock, how often should we try to get the lock
    # Default: 50
    # Unit: milliseconds
    retry_interval: 100

    # The client section documentation: https://github.com/roadrunner-server/docs/blob/cfd26f23e442fb3fe98533a05c99a9d0bb89e553/kv/redis.md
    client:
        addrs:
            - "${REDIS_HOST:-127.0.0.1}:${REDIS_PORT:-6379}"
        username: "${REDIS_USERNAME}"
        password: "${REDIS_PASSWORD}"
        db: "${REDIS_DB:-0}"
        #...
```

## PHP Implementation
```php
namespace App;

use DateInterval;
use RoadRunner\Lock\DTO\V1BETA1\Request;
use RoadRunner\Lock\DTO\V1BETA1\Response;
use RoadRunner\Lock\LockIdGeneratorInterface;
use RoadRunner\Lock\LockInterface;
use Spiral\Goridge\RPC\Codec\ProtobufCodec;
use Spiral\Goridge\RPC\RPCInterface;

/**
 * @see \RoadRunner\Lock\Lock
 */
final class RedisLockService implements LockInterface
{
    /**
     * @var RPCInterface
     */
    private RPCInterface $rpc;

    /**
     * @param RPCInterface $rpc
     * @param LockIdGeneratorInterface $identityGenerator
     */
    public function __construct(
        RPCInterface $rpc,
        private readonly LockIdGeneratorInterface $identityGenerator,
    ) {
        $this->rpc = $rpc->withCodec(new ProtobufCodec());
    }

    /**
     * @inheritDoc
     */
    public function lock(
        string $resource,
        ?string $id = null,
        int|float|DateInterval $ttl = 0,
        int|float|DateInterval $waitTTL = 0,
    ): false|string {
        $request = new Request();
        $request->setResource($resource);
        $request->setId($id ??= $this->identityGenerator->generate());
        $request->setTtl($this->convertTimeToMicroseconds($ttl));
        $request->setWait($this->convertTimeToMicroseconds($waitTTL));

        $response = $this->call('redislock.Lock', $request);

        return $response->getOk() ? $id : false;
    }

    /**
     * @inheritDoc
     */
    public function lockRead(
        string $resource,
        ?string $id = null,
        int|float|DateInterval $ttl = 0,
        int|float|DateInterval $waitTTL = 0,
    ): false|string {
        $request = new Request();
        $request->setResource($resource);
        $request->setId($id ??= $this->identityGenerator->generate());
        $request->setTtl($this->convertTimeToMicroseconds($ttl));
        $request->setWait($this->convertTimeToMicroseconds($waitTTL));

        $response = $this->call('redislock.LockRead', $request);

        return $response->getOk() ? $id : false;
    }

    /**
     * @inheritDoc
     */
    public function release(string $resource, string $id): bool
    {
        $request = new Request();
        $request->setResource($resource);
        $request->setId($id);

        $response = $this->call('redislock.Release', $request);

        return $response->getOk();
    }

    /**
     * @inheritDoc
     */
    public function forceRelease(string $resource): bool
    {
        $request = new Request();
        $request->setResource($resource);

        $response = $this->call('redislock.ForceRelease', $request);

        return $response->getOk();
    }

    /**
     * @inheritDoc
     */
    public function exists(string $resource, ?string $id = null): bool
    {
        $request = new Request();
        $request->setResource($resource);
        $request->setId($id ?? '*');

        $response = $this->call('redislock.Exists', $request);

        return $response->getOk();
    }

    /**
     * @inheritDoc
     */
    public function updateTTL(string $resource, string $id, int|float|DateInterval $ttl): bool
    {
        $request = new Request();
        $request->setResource($resource);
        $request->setId($id);
        $request->setTtl($this->convertTimeToMicroseconds($ttl));

        $response = $this->call('redislock.UpdateTTL', $request);

        return $response->getOk();
    }

    /**
     * @param int|float|DateInterval $ttl
     * @return int
     */
    private function convertTimeToMicroseconds(int|float|DateInterval $ttl): int
    {
        if ($ttl instanceof DateInterval) {
            return (int) \round((int)$ttl->format('%s') * 1_000_000);
        }

        \assert($ttl >= 0, 'TTL must be positive');

        return (int) \round($ttl * 1_000_000);
    }

    /**
     * @param string $method
     * @param Request $request
     * @return Response
     */
    private function call(string $method, Request $request): Response
    {
        $response = $this->rpc->call($method, $request, Response::class);
        \assert($response instanceof Response);

        return $response;
    }
}
```
