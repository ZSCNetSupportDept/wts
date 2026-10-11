// Optional local integration harness. Install embedded-postgres in a temporary directory.
// EMBEDDED_POSTGRES_MODULE must point to its dist/index.js; GO_BINARY defaults to go.
import { readFile, mkdtemp } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { spawn } from 'node:child_process';
import { createServer } from 'node:net';

const modulePath = process.env.EMBEDDED_POSTGRES_MODULE;
const { default: EmbeddedPostgres } = await import(modulePath ? pathToFileURL(modulePath).href : 'embedded-postgres');
const folder = await mkdtemp(join(tmpdir(),'wts-scheduler-test-'));
const port = await new Promise((resolve,reject) => {
    const server=createServer();server.on('error',reject);
    server.listen(0,'127.0.0.1',()=>{const value=server.address().port;server.close(()=>resolve(value));});
});
const database = new EmbeddedPostgres({databaseDir:join(folder,'db'),user:'postgres',password:'local-test-only',
    port,persistent:true,postgresFlags:['-h','127.0.0.1'],onLog:()=>{},onError:()=>{}});
const name='wts_scheduler_test';
let client;
try {
    await database.initialise();await database.start();await database.createDatabase(name);
    client=database.getPgClient(name,'127.0.0.1');await client.connect();
    await client.query("CREATE ROLE app LOGIN PASSWORD 'local-test-only'");
    let schema=await readFile(new URL('../src/db/wts3_v0.3.125-script_public.sql',import.meta.url),'utf8');
    schema=schema.replace(/CREATE DATABASE zsc[\s\S]*?;/,'').replaceAll('ON DATABASE zsc',`ON DATABASE ${name}`);
    await client.query(schema);
    await client.query(await readFile(new URL('../src/db/migrations/001_scheduler.sql',import.meta.url),'utf8'));
    await client.query(await readFile(new URL('../src/db/migrations/003_scheduler_notifications.sql',import.meta.url),'utf8'));
    await client.query(await readFile(new URL('../src/db/migrations/004_weekly_schedule.sql',import.meta.url),'utf8'));
    await client.query(await readFile(new URL('../src/db/migrations/005_remove_scheduler_notifications.sql',import.meta.url),'utf8'));
    await client.query(await readFile(new URL('../src/db/migrations/006_monthly_schedule.sql',import.meta.url),'utf8'));
    const env={...process.env,WTS_SCHEDULER_TEST_URL:`postgres://app:local-test-only@127.0.0.1:${port}/${name}?sslmode=disable`,
        WTS_SCHEDULER_TEST_ADMIN_URL:`postgres://postgres:local-test-only@127.0.0.1:${port}/${name}?sslmode=disable`};
    const code=await new Promise((resolve,reject)=>{
        const child=spawn(process.env.GO_BINARY||'go',['test','./...','-count=1'],{cwd:fileURLToPath(new URL('../src/',import.meta.url)),env,stdio:'inherit',windowsHide:true});
        child.on('error',reject);child.on('exit',resolve);
    });
    process.exitCode=code||0;
} finally {
    if(client) await client.end();
    await database.stop();
    // Retain the isolated test data for diagnostics; never remove computed directories here.
}
