import express from 'express';
import requireLogin from '../middleware/requireLogin';

let list = [
  {
    name: 'standby-tom',
    phase: 'Ready',
    clusterIp: '10.59.53.197',
    gmtCreate: '2020-06-18 15:53:05',
    capacity: 2048,
    secret: '123',
    dc: 'hz',
    env: 'production',
    kind: 'standby',
  },
  {
    name: 'cluster-jack',
    phase: 'Ready',
    clusterIp: '10.59.115.198',
    gmtCreate: '2020-06-18 14:37:00',
    capacity: 131072,
    secret: '123',
    dc: 'hz',
    env: 'production',
    kind: 'cluster',
  },
];

const router = express.Router();

router.get('/api/v1alpha2/redis', requireLogin, (req, res) => {
  res.json(list);
});

router.post('/api/v1alpha2/redis', requireLogin, (req, res) => {
  const ins = req.body;
  if (!ins.name || !ins.env || !ins.dc || !ins.kind) {
    res.sendStatus(400);
  } else if (list.find(o => o.name == ins.name)) {
    res.sendStatus(409);
  } else {
    list.push(ins);
    res.send({ success: true });
  }
});

router.put('/api/v1alpha2/redis', requireLogin, (req, res) => {
  const ins = req.body;
  if (!ins.name || !ins.env || !ins.dc || !ins.kind) {
    res.sendStatus(400);
  } else if (!list.find(o => o.name == ins.name)) {
    res.sendStatus(404);
  } else {
    list = list.filter(o => o.name !== ins.name);
    list.push(ins);
    res.send({ success: true });
  }
});

router.delete('/api/v1alpha2/redis', requireLogin, (req, res) => {
  const ins = req.body;
  // res.status(500);
  // res.send({ message: 'invalid', success: false });
  if (!ins.name) {
    res.sendStatus(400);
  } else if (!list.find(o => o.name == ins.name)) {
    res.send({ success: true });
  } else {
    list = list.filter(o => o.name !== ins.name);
    res.send({ success: true });
  }
});

export default router;
