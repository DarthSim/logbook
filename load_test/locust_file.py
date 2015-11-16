from locust import HttpLocust, TaskSet, task
import random

def app_name():
  return "app" + str(random.randint(1,5))

def tag_name():
  return "tag" + str(random.randint(1,1000))

def created_at():
  return "2015-01-02T0" + str(random.randint(0,9)) + ":" + str(random.randint(10,59)) + ":" + str(random.randint(10,59)) + ".123+06:00"

class UserBehavior(TaskSet):
  @task(10)
  def put_log(self):
    self.client.post(
      "/"+app_name()+"/put",
      {
        "level": 3,
        "message": "Lorem ipsum dolor",
        "tags": ",".join([tag_name(), tag_name(), tag_name()]),
        "created_at": created_at()
      },
      name="Put log"
    )

  @task(1)
  def get_log(self):
    self.client.get(
      "/"+app_name()+"/get?level=3&tags="+tag_name()+","+tag_name()+"&start_time=2015-01-01&end_time=2015-01-03",
      name="Get logs"
    )

class WebsiteUser(HttpLocust):
  host = "http://user:password@127.0.0.1:11610"
  task_set = UserBehavior
  min_wait=5000
  max_wait=9000
