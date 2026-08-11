import * as duckdb from "@duckdb/duckdb-wasm";

const bundles = duckdb.getJsDelivrBundles();

export async function createBrowserDatabase(): Promise<duckdb.AsyncDuckDB> {
  const bundle = await duckdb.selectBundle(bundles);
  const workerUrl = URL.createObjectURL(
    new Blob([`importScripts("${bundle.mainWorker!}");`], { type: "text/javascript" }),
  );
  const worker = new Worker(workerUrl);
  const database = new duckdb.AsyncDuckDB(new duckdb.VoidLogger(), worker);

  try {
    await database.instantiate(bundle.mainModule, bundle.pthreadWorker);
    return database;
  } catch (error) {
    await database.terminate();
    throw error;
  } finally {
    URL.revokeObjectURL(workerUrl);
  }
}

export const DuckDBDataProtocol = duckdb.DuckDBDataProtocol;
