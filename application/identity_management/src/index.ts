import dotenv from 'dotenv';
import express from 'express';
import { prepareContext, createUser, userExists } from './manageUser';
import enrollAdmin from './enrollAdmin';

dotenv.config();
const port = process.env.EXPRESS_PORT || 8000;

const ledgers: Array<string> = ['sla', 'vru', 'parts', 'sla2.0'];

enrollAdmin(1, ledgers[0]);
enrollAdmin(2, ledgers[1]);
enrollAdmin(3, ledgers[2]);
enrollAdmin(4, ledgers[3]);

prepareContext(1, ledgers[0]);
prepareContext(2, ledgers[1]);
prepareContext(3, ledgers[2]);
prepareContext(4, ledgers[3]);

const app = express();
app.use(express.json()); // for parsing application/json

app.post('/create', async (req, res) => {
  const { username, org } = req.body;
  const result = await createUser(username, org, ledgers[org - 1]);
  if (result.error === undefined) {
    console.debug(`Created  user: ${username} in wallet ${ledgers[org - 1]}`);
    return res.json({ success: true, data: result });
  }
  return res.json({ success: false, error: result.error });
});

app.post('/exists', async (req, res) => {
  const { cert } = req.body;
  const { found, org, username } = await userExists(cert.replaceAll('\n', ''));
  if (found) {
    console.debug(`User ${username} found on organisation ${org}`);
  } else {
    console.debug('User not found');
  }

  return res.json({
    success: true, exists: found, organisation: org, username,
  });
});

app.listen(port, () => {
  console.debug(`Identity management app listening on port ${port}`);
});
