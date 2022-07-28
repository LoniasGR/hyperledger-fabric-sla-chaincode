import dotenv from 'dotenv';
import express from 'express';
import { prepareContext, createUser } from './createUser';
import enrollAdmin from './enrollAdmin';

dotenv.config();
const port = process.env.EXPRESS_PORT || 3000;

const ledgers: Array<string> = ["sla", "vru", "parts"];

enrollAdmin(1, ledgers[0]);
enrollAdmin(2, ledgers[1]);
enrollAdmin(3, ledgers[2]);

prepareContext(1, ledgers[0]);
prepareContext(2, ledgers[1]);
prepareContext(3, ledgers[2]);

const app = express();
app.use(express.json()); // for parsing application/json

app.post('/', async (req, res) => {
  const { username, org } = req.body;
  const result = await createUser(username, org, ledgers[org-1]);
  if (result.error === undefined) {
    console.debug(`Created  user: ${username} in wallet ${ledgers[org-1]}`);
    return res.json({ success: true, data: result });
  }
  return res.json({ success: false, error: result.error });
});

app.listen(port, () => {
  console.debug(`Example app listening on port ${port}`);
});
