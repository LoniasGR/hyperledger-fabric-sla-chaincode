import dotenv from 'dotenv';
import express from 'express';
import { prepareContext, createUser } from './createUser';
import enrollAdmin from './enrollAdmin';

dotenv.config();
const port = process.env.EXPRESS_PORT || 3000;

enrollAdmin();
prepareContext();

const app = express();
app.use(express.json()); // for parsing application/json

app.post('/', async (req, res) => {
  const { username } = req.body;
  const result = await createUser(username);
  if (result.error === undefined) {
    console.debug(`Created  user: ${username}`);
    return res.json({ success: true, data: result });
  }
  return res.json({ success: false, error: result.error });
});

app.listen(port, () => {
  console.debug(`Example app listening on port ${port}`);
});
